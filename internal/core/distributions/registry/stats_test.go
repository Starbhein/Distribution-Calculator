package registry

import (
	"testing"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
)

// Tests for the registry stats helpers (PR4b-1 scope, spec §3, design
// §1.3): TheoreticalStats and Probabilities delegate to the distribution
// structs' own methods via Construct — no inline formulas. Every case
// asserts equality with the struct methods for the same parameters.

func TestTheoreticalStatsDelegatesToStruct(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		params      []float64 // UI-facing order (binomial is (p, n))
	}{
		{"normal", "Normal", []float64{10, 2}},
		{"exponential lambda", "Exponencial (λ)", []float64{2}},
		{"exponential beta", "Exponencial (β)", []float64{2}},
		{"uniform", "Uniforme continua", []float64{2, 4}},
		{"bernoulli", "Bernoulli", []float64{0.3}},
		{"binomial", "Binomial", []float64{0.5, 10}},
		{"geometric", "Geométrica", []float64{0.25}},
		{"poisson", "Poisson", []float64{4}},
		{"hypergeometric", "Hypergeométrica", []float64{12, 3, 4}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := ByName(tc.displayName)
			if !ok {
				t.Fatalf("ByName(%q) returned ok=false", tc.displayName)
			}
			avg, variance, stdDev, err := TheoreticalStats(spec, tc.params)
			if err != nil {
				t.Fatalf("TheoreticalStats(%v) returned error: %v", tc.params, err)
			}
			d, err := spec.Construct(tc.params)
			if err != nil {
				t.Fatalf("Construct(%v) returned error: %v", tc.params, err)
			}
			evalFloats(t, avg, d.Avg())
			evalFloats(t, variance, d.Variance())
			evalFloats(t, stdDev, d.StdDev())
		})
	}
}

func TestTheoreticalStatsInvalidParams(t *testing.T) {
	// The helper validates through Construct (design §7 convergence):
	// invalid params are an error, not silent garbage.
	spec, _ := ByName("Normal")
	if _, _, _, err := TheoreticalStats(spec, []float64{10, 0}); err == nil {
		t.Error("TheoreticalStats with sigma=0 expected error, got nil")
	}
}

func TestProbabilitiesDelegatesToStruct(t *testing.T) {
	t.Run("discrete", func(t *testing.T) {
		cases := []struct {
			name        string
			displayName string
			params      []float64
			k           int
		}{
			{"bernoulli", "Bernoulli", []float64{0.3}, 1},
			{"binomial", "Binomial", []float64{0.5, 10}, 5},
			{"geometric", "Geométrica", []float64{0.25}, 3},
			{"poisson", "Poisson", []float64{4}, 3},
			{"hypergeometric", "Hypergeométrica", []float64{12, 3, 4}, 0},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				spec, ok := ByName(tc.displayName)
				if !ok {
					t.Fatalf("ByName(%q) returned ok=false", tc.displayName)
				}
				x := float64(tc.k)
				px, ple, pgt, err := Probabilities(spec, tc.params, x)
				if err != nil {
					t.Fatalf("Probabilities(%v, %v) returned error: %v", tc.params, x, err)
				}
				d, err := spec.Construct(tc.params)
				if err != nil {
					t.Fatalf("Construct(%v) returned error: %v", tc.params, err)
				}
				dd, ok := d.(distributions.DiscreteDistribution)
				if !ok {
					t.Fatalf("Construct did not return a DiscreteDistribution: %T", d)
				}
				pmfWant, err := dd.PMF(tc.k)
				if err != nil {
					t.Fatalf("PMF(%d) returned error: %v", tc.k, err)
				}
				cdfWant, err := dd.CDF(tc.k)
				if err != nil {
					t.Fatalf("CDF(%d) returned error: %v", tc.k, err)
				}
				evalFloats(t, px, pmfWant)
				evalFloats(t, ple, cdfWant)
				evalFloats(t, pgt, 1-cdfWant)
			})
		}
	})

	t.Run("continuous", func(t *testing.T) {
		cases := []struct {
			name        string
			displayName string
			params      []float64
			x           float64
		}{
			{"normal", "Normal", []float64{10, 2}, 10},
			{"exponential lambda", "Exponencial (λ)", []float64{2}, 0.5},
			{"exponential beta", "Exponencial (β)", []float64{2}, 2},
			{"uniform", "Uniforme continua", []float64{2, 4}, 3},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				spec, ok := ByName(tc.displayName)
				if !ok {
					t.Fatalf("ByName(%q) returned ok=false", tc.displayName)
				}
				px, ple, pgt, err := Probabilities(spec, tc.params, tc.x)
				if err != nil {
					t.Fatalf("Probabilities(%v, %v) returned error: %v", tc.params, tc.x, err)
				}
				d, err := spec.Construct(tc.params)
				if err != nil {
					t.Fatalf("Construct(%v) returned error: %v", tc.params, err)
				}
				cd, ok := d.(distributions.ContinuousDistribution)
				if !ok {
					t.Fatalf("Construct did not return a ContinuousDistribution: %T", d)
				}
				pdfWant, err := cd.PDF(tc.x)
				if err != nil {
					t.Fatalf("PDF(%v) returned error: %v", tc.x, err)
				}
				cdfWant, err := cd.CDF(tc.x)
				if err != nil {
					t.Fatalf("CDF(%v) returned error: %v", tc.x, err)
				}
				evalFloats(t, px, pdfWant)
				evalFloats(t, ple, cdfWant)
				evalFloats(t, pgt, 1-cdfWant)
			})
		}
	})
}

func TestProbabilitiesDiscreteRoundsX(t *testing.T) {
	// The UI passes x as a float64; the discrete path rounds to the
	// nearest integer, preserving the old ComputeProbabilities behavior.
	spec, _ := ByName("Binomial")
	pxRounded, _, _, err := Probabilities(spec, []float64{0.5, 10}, 4.6)
	if err != nil {
		t.Fatalf("Probabilities returned error: %v", err)
	}
	pxExact, _, _, err := Probabilities(spec, []float64{0.5, 10}, 5)
	if err != nil {
		t.Fatalf("Probabilities returned error: %v", err)
	}
	evalFloats(t, pxRounded, pxExact)
}
