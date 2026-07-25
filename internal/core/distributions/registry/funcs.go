package registry

import (
	"math"

	"github.com/Starbhein/DistCalc/internal/core/distmath"
	"github.com/Starbhein/DistCalc/internal/core/distributions"
)

// Render-path helpers (PR4c scope, spec §5, design §1.3/§3.3). They replace
// the chart.go:149,324 and plot.go:249,295 switches: the distribution is
// constructed ONCE per render (never per bar/bin), and the discrete row is
// backed by one distmath O(range) pass for binomial/poisson/hypergeometric.
// Bernoulli/Geometric keep their cheap closed forms. Displayed values are
// unchanged: row-vs-closed-form drift is ~1 ulp, invisible at the %.4f/%.3f
// formatting every render site uses.

// maxPMFRowLen caps the materialized PMF row length at ~1M entries (~8 MiB
// of float64s). Validators only enforce n>0/lambda>0 and the UI CharLimit
// allows huge values, so without this gate a full-support row could allocate
// gigabytes (binomial n=1e9 ~ 8 GiB) before the chart draws (review
// RES-1/RES-2/REL-1). Beyond the cap the closed-form struct PMF is used —
// the render path only iterates the empirical data range anyway. It is a
// var (not const) so tests can lower it to exercise the fallback cheaply.
var maxPMFRowLen = 1 << 20

// PMFFunc returns the theoretical PMF evaluator for a discrete spec, or nil
// when the spec is continuous or the params fail validation. Out-of-support
// k yields 0, matching the old error-ignoring per-bar calls.
func PMFFunc(spec Spec, params []float64) func(int) float64 {
	if spec.Discrete == false {
		return nil
	}
	d, err := spec.Construct(params)
	if err != nil {
		return nil
	}
	dd, ok := d.(distributions.DiscreteDistribution)
	if !ok {
		return nil
	}
	switch spec.ID {
	case IDBinomial:
		// Engine order (n, p) comes from the SAME single-sourced binding
		// Construct/NewSampler use — the (p, n) → (n, p) swap is never
		// re-derived at a call site.
		n, p, err := bindBinomial(params)
		if err != nil {
			return nil
		}
		if n+1 > maxPMFRowLen {
			// Oversized support: closed-form fallback (same pattern as
			// the Poisson tail below) instead of a gigabyte-scale row.
			return func(k int) float64 {
				if k < 0 || k > n {
					return 0
				}
				v, _ := dd.PMF(k)
				return v
			}
		}
		row := distmath.BinomialPMFRow(n, p)
		return func(k int) float64 {
			if k < 0 || k >= len(row) {
				return 0
			}
			return row[k]
		}
	case IDPoisson:
		lambda := params[0]
		// Prospective row length is int(lambda + 4*sqrt(lambda)) + 2
		// (distmath.PoissonPMFRow extent rule); gate before materializing.
		if int(lambda+4.0*math.Sqrt(lambda))+2 > maxPMFRowLen {
			return func(k int) float64 {
				if k < 0 {
					return 0
				}
				v, _ := dd.PMF(k)
				return v
			}
		}
		row := distmath.PoissonPMFRow(lambda)
		return func(k int) float64 {
			if k < 0 {
				return 0
			}
			if k < len(row) {
				return row[k]
			}
			// Tail beyond the row extent: closed-form fallback keeps
			// tiny-but-displayable masses byte-identical (design §3.3).
			v, _ := dd.PMF(k)
			return v
		}
	case IDHypergeometric:
		// The ONLY UI-order (N,M,n) → engine-order (m,nsample,n) mapping,
		// reused from the spec's named binding (design §4).
		hp, err := bindHypergeometric(params)
		if err != nil {
			return nil
		}
		// Prospective support extent [startK, maxK] = [max(0, n+M-N),
		// min(M, n)]: gate before materializing support+1 floats.
		startK := max(0, hp.Sample+hp.Successes-hp.Population)
		maxK := min(hp.Successes, hp.Sample)
		if maxK-startK+1 > maxPMFRowLen {
			return func(k int) float64 {
				if k < startK || k > maxK {
					return 0
				}
				v, _ := dd.PMF(k)
				return v
			}
		}
		row, startK, maxK := distmath.HypergeometricPMFRow(hp.Successes, hp.Sample, hp.Population)
		return func(k int) float64 {
			if k < startK || k > maxK {
				return 0
			}
			return row[k-startK]
		}
	default:
		// Bernoulli/Geometric: cheap closed forms, constructed once above.
		return func(k int) float64 {
			v, _ := dd.PMF(k)
			return v
		}
	}
}

// PDFFunc returns the theoretical PDF evaluator for a continuous spec, or
// nil when the spec is discrete or the params fail validation (the old
// buildPDFFunc nil semantics — no theoretical curve is drawn).
func PDFFunc(spec Spec, params []float64) func(float64) float64 {
	if spec.Discrete {
		return nil
	}
	d, err := spec.Construct(params)
	if err != nil {
		return nil
	}
	cd, ok := d.(distributions.ContinuousDistribution)
	if !ok {
		return nil
	}
	return func(x float64) float64 {
		v, _ := cd.PDF(x)
		return v
	}
}
