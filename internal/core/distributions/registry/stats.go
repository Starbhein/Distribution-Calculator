package registry

import (
	"errors"
	"math"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
)

// Stats helpers (design §1.3): they replace the theoretical.go:19 and
// theoretical.go:159 switches by delegating to the distribution structs'
// own methods via Construct — no inline formulas anywhere (spec §3).

// TheoreticalStats returns the theoretical Avg/Variance/StdDev for the
// spec's distribution, obtained from the struct methods. Params are in
// UI-facing order and are validated through Construct (design §7).
func TheoreticalStats(spec Spec, params []float64) (avg, variance, stdDev float64, err error) {
	d, err := spec.Construct(params)
	if err != nil {
		return 0, 0, 0, err
	}
	return d.Avg(), d.Variance(), d.StdDev(), nil
}

// Probabilities returns P(X=x), P(X≤x) and P(X>x) for the spec's
// distribution at x, obtained from the struct methods. Discrete specs
// round x to the nearest integer (preserving the old
// ComputeProbabilities behavior); continuous specs evaluate the PDF.
func Probabilities(spec Spec, params []float64, x float64) (px, ple, pgt float64, err error) {
	d, err := spec.Construct(params)
	if err != nil {
		return 0, 0, 0, err
	}
	if spec.Discrete {
		dd, ok := d.(distributions.DiscreteDistribution)
		if !ok {
			return 0, 0, 0, errors.New("registry: discrete spec did not construct a DiscreteDistribution")
		}
		k := int(math.Round(x))
		pmf, err := dd.PMF(k)
		if err != nil {
			return 0, 0, 0, err
		}
		cdf, err := dd.CDF(k)
		if err != nil {
			return 0, 0, 0, err
		}
		return pmf, cdf, 1 - cdf, nil
	}
	cd, ok := d.(distributions.ContinuousDistribution)
	if !ok {
		return 0, 0, 0, errors.New("registry: continuous spec did not construct a ContinuousDistribution")
	}
	pdf, err := cd.PDF(x)
	if err != nil {
		return 0, 0, 0, err
	}
	cdf, err := cd.CDF(x)
	if err != nil {
		return 0, 0, 0, err
	}
	return pdf, cdf, 1 - cdf, nil
}
