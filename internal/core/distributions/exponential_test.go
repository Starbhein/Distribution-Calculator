package distributions

import "testing"

// Gate-0 characterization tests (spec §2, design §6.2).
// These pin the CURRENT behavior of both exponential parameterizations
// before any refactor or rename. All values are derived from the
// unmodified implementation.

func TestExponentialLambdaCharacterization(t *testing.T) {
	t.Run("lambda=2 pinned point values", func(t *testing.T) {
		exp, err := NewExponentialLambda(2)
		if err != nil {
			t.Fatalf("NewExponentialLambda(2) returned error: %v", err)
		}

		pdf0, err := exp.PDF(0)
		if err != nil {
			t.Fatalf("PDF(0) returned error: %v", err)
		}
		evalFloats(t, pdf0, 2)

		pdf05, err := exp.PDF(0.5)
		if err != nil {
			t.Fatalf("PDF(0.5) returned error: %v", err)
		}
		evalFloats(t, pdf05, 0.7357588823428847)

		cdf05, err := exp.CDF(0.5)
		if err != nil {
			t.Fatalf("CDF(0.5) returned error: %v", err)
		}
		evalFloats(t, cdf05, 0.6321205588285577)

		evalFloats(t, exp.Avg(), 0.5)
		evalFloats(t, exp.Variance(), 0.25)
		evalFloats(t, exp.StdDev(), 0.5)
	})

	t.Run("constructor rejects negative lambda", func(t *testing.T) {
		if got, err := NewExponentialLambda(-1); err == nil || got != nil {
			t.Errorf("NewExponentialLambda(-1) = (%v, %v), want (nil, non-nil error)", got, err)
		}
	})

	t.Run("PDF and CDF reject negative x", func(t *testing.T) {
		exp, err := NewExponentialLambda(2)
		if err != nil {
			t.Fatalf("NewExponentialLambda(2) returned error: %v", err)
		}
		if _, err := exp.PDF(-0.5); err == nil {
			t.Error("PDF(-0.5) want non-nil error")
		}
		if _, err := exp.CDF(-0.5); err == nil {
			t.Error("CDF(-0.5) want non-nil error")
		}
	})
}

func TestExponentialBetaCharacterization(t *testing.T) {
	t.Run("beta=2 pinned point values", func(t *testing.T) {
		exp, err := NewExponentialBeta(2)
		if err != nil {
			t.Fatalf("NewExponentialBeta(2) returned error: %v", err)
		}

		// Quirk pinned as-is: PDF(0) returns 1 instead of the
		// mathematically correct 1/beta (exponential.go). This is a
		// characterization test pinning current behavior — do NOT
		// "fix" the implementation here.
		pdf0, err := exp.PDF(0)
		if err != nil {
			t.Fatalf("PDF(0) returned error: %v", err)
		}
		evalFloats(t, pdf0, 1)

		pdf2, err := exp.PDF(2)
		if err != nil {
			t.Fatalf("PDF(2) returned error: %v", err)
		}
		evalFloats(t, pdf2, 0.18393972058572117)

		cdf2, err := exp.CDF(2)
		if err != nil {
			t.Fatalf("CDF(2) returned error: %v", err)
		}
		evalFloats(t, cdf2, 0.6321205588285577)

		evalFloats(t, exp.Avg(), 2)
		evalFloats(t, exp.Variance(), 4)
		evalFloats(t, exp.StdDev(), 2)
	})

	t.Run("constructor rejects negative beta", func(t *testing.T) {
		if got, err := NewExponentialBeta(-1); err == nil || got != nil {
			t.Errorf("NewExponentialBeta(-1) = (%v, %v), want (nil, non-nil error)", got, err)
		}
	})

	t.Run("PDF and CDF reject negative x", func(t *testing.T) {
		exp, err := NewExponentialBeta(2)
		if err != nil {
			t.Fatalf("NewExponentialBeta(2) returned error: %v", err)
		}
		if _, err := exp.PDF(-0.5); err == nil {
			t.Error("PDF(-0.5) want non-nil error")
		}
		if _, err := exp.CDF(-0.5); err == nil {
			t.Error("CDF(-0.5) want non-nil error")
		}
	})
}

func TestExponentialLambdaBethaRelationship(t *testing.T) {
	// beta = 1/lambda semantics as currently implemented:
	// ExponentialLambda(l).PDF(x) == ExponentialBeta(1/l).PDF(x) for x > 0.
	lambdas := []float64{0.5, 1, 2, 3.7}
	xs := []float64{0.25, 1, 4}
	for _, lambda := range lambdas {
		el, err := NewExponentialLambda(lambda)
		if err != nil {
			t.Fatalf("NewExponentialLambda(%v) returned error: %v", lambda, err)
		}
		eb, err := NewExponentialBeta(1 / lambda)
		if err != nil {
			t.Fatalf("NewExponentialBeta(%v) returned error: %v", 1/lambda, err)
		}
		for _, x := range xs {
			gotLambda, err := el.PDF(x)
			if err != nil {
				t.Fatalf("ExponentialLambda(%v).PDF(%v) returned error: %v", lambda, x, err)
			}
			gotBetha, err := eb.PDF(x)
			if err != nil {
				t.Fatalf("ExponentialBeta(%v).PDF(%v) returned error: %v", 1/lambda, x, err)
			}
			evalFloats(t, gotLambda, gotBetha)
		}
	}
}
