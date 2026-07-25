package registry

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
)

// Tests for the render-path helpers (PR4c scope, spec §5, design §1.3/§3.3):
// PMFFunc and PDFFunc construct the distribution ONCE and back the discrete
// row with one distmath row pass for binomial/poisson/hypergeometric;
// Bernoulli/Geometric keep their cheap closed forms. Values must match the
// struct PMF/PDF within 1e-12, and the %.4f/%.3f formatting used by every
// render site must be byte-identical.

func TestPMFFuncMatchesStructPMF(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		params      []float64 // UI-facing order (binomial is (p, n))
		kMax        int
	}{
		{"bernoulli", "Bernoulli", []float64{0.3}, 1},
		{"binomial (p,n) UI order", "Binomial", []float64{0.25, 40}, 40},
		{"geometric", "Geométrica", []float64{0.25}, 20},
		{"poisson", "Poisson", []float64{4}, 25},
		{"hypergeometric", "Hypergeométrica", []float64{12, 3, 4}, 12},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := ByName(tc.displayName)
			if !ok {
				t.Fatalf("ByName(%q) returned ok=false", tc.displayName)
			}
			pmfFn := PMFFunc(spec, tc.params)
			if pmfFn == nil {
				t.Fatalf("PMFFunc(%q, %v) returned nil", tc.displayName, tc.params)
			}
			d, err := spec.Construct(tc.params)
			if err != nil {
				t.Fatalf("Construct(%v) returned error: %v", tc.params, err)
			}
			dd := d.(distributions.DiscreteDistribution)
			for k := -2; k <= tc.kMax+2; k++ {
				want, werr := dd.PMF(k)
				if werr != nil {
					want = 0 // render sites show 0 outside the support
				}
				evalFloats(t, pmfFn(k), want)
				// Display-identical proof: the %.4f format used by every
				// render site absorbs the ~1 ulp row-vs-closed-form drift.
				if got, exp := fmt.Sprintf("%.4f", pmfFn(k)), fmt.Sprintf("%.4f", want); got != exp {
					t.Errorf("k=%d: formatted PMF %s != %s", k, got, exp)
				}
			}
		})
	}
}

// TestPMFFuncPoissonTailFallback pins the out-of-row behavior: for k beyond
// the row extent (lambda + 4*sigma + 1) the helper falls back to the struct
// closed form, so tiny-but-displayable tail masses (PMF(5; 0.5) renders as
// 0.0002) stay byte-identical to the pre-refactor per-bar calls.
func TestPMFFuncPoissonTailFallback(t *testing.T) {
	spec, _ := ByName("Poisson")
	params := []float64{0.5}
	pmfFn := PMFFunc(spec, params)
	if pmfFn == nil {
		t.Fatal("PMFFunc(Poisson, [0.5]) returned nil")
	}
	p, _ := distributions.NewPoisson(0.5)
	for k := 0; k <= 8; k++ {
		want, _ := p.PMF(k)
		evalFloats(t, pmfFn(k), want)
		if got, exp := fmt.Sprintf("%.4f", pmfFn(k)), fmt.Sprintf("%.4f", want); got != exp {
			t.Errorf("k=%d: formatted PMF %s != %s", k, got, exp)
		}
	}
}

func TestPMFFuncNilForContinuousOrInvalid(t *testing.T) {
	continuous, _ := ByName("Normal")
	if fn := PMFFunc(continuous, []float64{10, 2}); fn != nil {
		t.Error("PMFFunc on a continuous spec returned non-nil")
	}
	binomial, _ := ByName("Binomial")
	if fn := PMFFunc(binomial, []float64{1.5, 10}); fn != nil {
		t.Error("PMFFunc with invalid params (p>1) returned non-nil")
	}
}

func TestPDFFuncMatchesStructPDF(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		params      []float64
		lo, hi      float64
	}{
		{"normal", "Normal", []float64{10, 2}, 4, 16},
		{"exponential lambda", "Exponencial (λ)", []float64{2}, 0, 3},
		{"exponential beta", "Exponencial (β)", []float64{2}, 0, 6},
		{"uniform", "Uniforme continua", []float64{2, 4}, 1, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := ByName(tc.displayName)
			if !ok {
				t.Fatalf("ByName(%q) returned ok=false", tc.displayName)
			}
			pdfFn := PDFFunc(spec, tc.params)
			if pdfFn == nil {
				t.Fatalf("PDFFunc(%q, %v) returned nil", tc.displayName, tc.params)
			}
			d, err := spec.Construct(tc.params)
			if err != nil {
				t.Fatalf("Construct(%v) returned error: %v", tc.params, err)
			}
			cd := d.(distributions.ContinuousDistribution)
			for i := 0; i <= 20; i++ {
				x := tc.lo + float64(i)*(tc.hi-tc.lo)/20
				want, werr := cd.PDF(x)
				if werr != nil {
					want = 0
				}
				evalFloats(t, pdfFn(x), want)
				if got, exp := fmt.Sprintf("%.3f", pdfFn(x)), fmt.Sprintf("%.3f", want); got != exp {
					t.Errorf("x=%v: formatted PDF %s != %s", x, got, exp)
				}
			}
		})
	}
}

func TestPDFFuncNilForDiscreteOrInvalid(t *testing.T) {
	discrete, _ := ByName("Poisson")
	if fn := PDFFunc(discrete, []float64{4}); fn != nil {
		t.Error("PDFFunc on a discrete spec returned non-nil")
	}
	normal, _ := ByName("Normal")
	if fn := PDFFunc(normal, []float64{10, 0}); fn != nil {
		t.Error("PDFFunc with invalid params (sigma=0) returned non-nil")
	}
}

// Tests for the PMF row materialization cap (review RES-1/RES-2/REL-1):
// PMFFunc must never materialize a full-support row whose length exceeds
// maxPMFRowLen — validators accept unbounded n/lambda, so a huge-but-valid
// param would otherwise allocate gigabytes before the chart draws. Beyond
// the cap the helper falls back to the constructed struct's closed-form
// PMF, exactly like the existing Poisson tail fallback.

// TestPMFFuncRowCapBoundsAllocation proves the gate: one entry past the cap
// the row path would allocate (cap+11)*8 bytes (~8 MiB); the fallback must
// stay under 1 MiB while returning struct-identical values.
func TestPMFFuncRowCapBoundsAllocation(t *testing.T) {
	spec, _ := ByName("Binomial")
	n := maxPMFRowLen + 10
	params := []float64{0.5, float64(n)}
	_ = PMFFunc(spec, params) // warmup: stabilizes allocs before measuring
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	pmfFn := PMFFunc(spec, params)
	runtime.ReadMemStats(&after)
	if pmfFn == nil {
		t.Fatalf("PMFFunc(Binomial, %v) returned nil", params)
	}
	if got := after.TotalAlloc - before.TotalAlloc; got > 1<<20 {
		t.Errorf("PMFFunc materialized %d bytes for n=%d (cap %d); want bounded closed-form fallback",
			got, n, maxPMFRowLen)
	}
	d, err := spec.Construct(params)
	if err != nil {
		t.Fatalf("Construct(%v) returned error: %v", params, err)
	}
	dd := d.(distributions.DiscreteDistribution)
	for _, k := range []int{-1, 0, 1, n / 2, n - 1, n, n + 1} {
		want, werr := dd.PMF(k)
		if werr != nil {
			want = 0
		}
		evalFloats(t, pmfFn(k), want)
	}
}

// TestPMFFuncCapFallbackMatchesStructPMF lowers the cap so the cheap
// table-driven params exercise the fallback path, then checks the full
// displayed range against the struct PMF for all three row-backed IDs.
func TestPMFFuncCapFallbackMatchesStructPMF(t *testing.T) {
	orig := maxPMFRowLen
	maxPMFRowLen = 16
	defer func() { maxPMFRowLen = orig }()
	cases := []struct {
		name        string
		displayName string
		params      []float64
		kMax        int
	}{
		{"binomial", "Binomial", []float64{0.25, 40}, 40},
		{"poisson", "Poisson", []float64{25}, 60},
		{"hypergeometric", "Hypergeométrica", []float64{12, 3, 4}, 12},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, _ := ByName(tc.displayName)
			pmfFn := PMFFunc(spec, tc.params)
			if pmfFn == nil {
				t.Fatalf("PMFFunc(%q, %v) returned nil", tc.displayName, tc.params)
			}
			d, err := spec.Construct(tc.params)
			if err != nil {
				t.Fatalf("Construct(%v) returned error: %v", tc.params, err)
			}
			dd := d.(distributions.DiscreteDistribution)
			for k := -2; k <= tc.kMax+2; k++ {
				want, werr := dd.PMF(k)
				if werr != nil {
					want = 0
				}
				evalFloats(t, pmfFn(k), want)
			}
		})
	}
}

// TestPMFFuncHugeParamsNoOOM kills the review OOM scenario: huge-but-valid
// params that pass the validators must return a working closed-form func
// without materializing multi-gigabyte rows (binomial n=1e9 ~ 8 GiB,
// poisson lambda=1e8 ~ 800 MB, hypergeometric support 5e8 ~ 4 GiB).
func TestPMFFuncHugeParamsNoOOM(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		params      []float64
		ks          []int
	}{
		{"binomial n=1e9", "Binomial", []float64{0.5, 1e9},
			[]int{-1, 0, 1, 499_999_999, 500_000_000, 999_999_999, 1_000_000_000, 1_000_000_001}},
		{"poisson lambda=1e8", "Poisson", []float64{1e8},
			[]int{-1, 0, 99_990_000, 100_000_000, 100_010_000}},
		{"hypergeometric N=1e9", "Hypergeométrica", []float64{1e9, 5e8, 5e8},
			[]int{-1, 249_999_999, 250_000_000, 250_000_001, 500_000_001}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, _ := ByName(tc.displayName)
			pmfFn := PMFFunc(spec, tc.params)
			if pmfFn == nil {
				t.Fatalf("PMFFunc(%q, %v) returned nil", tc.displayName, tc.params)
			}
			d, err := spec.Construct(tc.params)
			if err != nil {
				t.Fatalf("Construct(%v) returned error: %v", tc.params, err)
			}
			dd := d.(distributions.DiscreteDistribution)
			for _, k := range tc.ks {
				want, werr := dd.PMF(k)
				if werr != nil {
					want = 0
				}
				evalFloats(t, pmfFn(k), want)
			}
		})
	}
}
