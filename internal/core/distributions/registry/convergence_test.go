package registry

import (
	"testing"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
)

// Validation-convergence guard (design §7/§9, spec §3): THE single
// validation layer is Spec.Validate; Construct and NewSampler can never
// accept what Validate rejects, and NewHypergeometric carries the
// cross-param backstop for direct library consumers. A second UI-side
// validator no longer exists — this test fails loudly if the layers ever
// disagree again.

func TestValidationConvergence(t *testing.T) {
	t.Run("invalid params rejected through the whole unified path", func(t *testing.T) {
		cases := []struct {
			name        string
			displayName string
			params      []float64
			wantIndex   int
		}{
			{"normal sigma zero", "Normal", []float64{10, 0}, 1},
			{"normal arity", "Normal", []float64{10}, -1},
			{"exponential lambda zero", "Exponencial (λ)", []float64{0}, 0},
			{"exponential beta negative", "Exponencial (β)", []float64{-1}, 0},
			{"uniform a >= b", "Uniforme continua", []float64{4, 2}, 0},
			{"bernoulli p zero", "Bernoulli", []float64{0}, 0},
			{"binomial p above one", "Binomial", []float64{1.5, 10}, 0},
			{"binomial n zero", "Binomial", []float64{0.5, 0}, 1},
			{"geometric p above one", "Geométrica", []float64{2}, 0},
			{"poisson lambda negative", "Poisson", []float64{-1}, 0},
			{"hypergeometric N zero", "Hypergeométrica", []float64{0, 3, 4}, 0},
			{"hypergeometric M zero", "Hypergeométrica", []float64{12, 0, 4}, 1},
			{"hypergeometric n zero", "Hypergeométrica", []float64{12, 3, 0}, 2},
			{"hypergeometric M greater than N", "Hypergeométrica", []float64{12, 13, 4}, 1},
			{"hypergeometric n greater than N", "Hypergeométrica", []float64{12, 3, 13}, 2},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				spec, ok := ByName(tc.displayName)
				if !ok {
					t.Fatalf("ByName(%q) returned ok=false", tc.displayName)
				}
				idx, err := spec.Validate(tc.params)
				if err == nil {
					t.Fatalf("Validate(%v) expected error, got nil", tc.params)
				}
				if idx != tc.wantIndex {
					t.Errorf("Validate paramIndex: got %d, want %d (form-highlight shape)", idx, tc.wantIndex)
				}
				if _, err := spec.Construct(tc.params); err == nil {
					t.Errorf("Construct(%v) expected error, got nil — layer disagreement", tc.params)
				}
				if _, err := spec.NewSampler(tc.params); err == nil {
					t.Errorf("NewSampler(%v) expected error, got nil — layer disagreement", tc.params)
				}
			})
		}
	})

	t.Run("valid params accepted everywhere — no regression", func(t *testing.T) {
		// Spec §3 acceptance: no parameter set accepted today is rejected.
		cases := []struct {
			name        string
			displayName string
			params      []float64
		}{
			{"normal", "Normal", []float64{10, 2}},
			{"exponential lambda", "Exponencial (λ)", []float64{2}},
			{"exponential beta", "Exponencial (β)", []float64{2}},
			{"uniform", "Uniforme continua", []float64{2, 4}},
			{"bernoulli p one", "Bernoulli", []float64{1}},
			{"binomial p one", "Binomial", []float64{1, 10}},
			{"geometric", "Geométrica", []float64{0.25}},
			{"poisson", "Poisson", []float64{4}},
			{"hypergeometric M equals N", "Hypergeométrica", []float64{12, 12, 4}},
			{"hypergeometric n equals N", "Hypergeométrica", []float64{12, 3, 12}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				spec, ok := ByName(tc.displayName)
				if !ok {
					t.Fatalf("ByName(%q) returned ok=false", tc.displayName)
				}
				if idx, err := spec.Validate(tc.params); err != nil {
					t.Fatalf("Validate(%v) returned error at index %d: %v", tc.params, idx, err)
				}
				if _, err := spec.Construct(tc.params); err != nil {
					t.Errorf("Construct(%v) returned error: %v", tc.params, err)
				}
				if _, err := spec.NewSampler(tc.params); err != nil {
					t.Errorf("NewSampler(%v) returned error: %v", tc.params, err)
				}
			})
		}
	})

	t.Run("constructor backstop agrees with the registry layer", func(t *testing.T) {
		// design §7: NewHypergeometric rejects the same cross-param combos
		// (M>N, n>N) that the registry Validate rejects — constructor order
		// is (M, N, n).
		backstopCases := []struct {
			name      string
			successes int
			pop       int
			sample    int
		}{
			{"M greater than N", 13, 12, 4},
			{"n greater than N", 3, 12, 13},
		}
		for _, tc := range backstopCases {
			t.Run(tc.name, func(t *testing.T) {
				if _, err := distributions.NewHypergeometric(tc.successes, tc.pop, tc.sample); err == nil {
					t.Errorf("NewHypergeometric(M=%d, N=%d, n=%d) expected error, got nil — backstop missing",
						tc.successes, tc.pop, tc.sample)
				}
			})
		}
	})
}
