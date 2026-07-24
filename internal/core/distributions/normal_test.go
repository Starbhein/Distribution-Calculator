package distributions

import "testing"

// Gate-0 characterization tests (spec §2, design §6.1).
// These pin the CURRENT behavior of Normal before any refactor or rename.
// All values are derived from the unmodified implementation.

func TestNormalCharacterization(t *testing.T) {
	t.Run("N(0,1) pinned point values", func(t *testing.T) {
		normal, err := NewNormal(0, 1)
		if err != nil {
			t.Fatalf("NewNormal(0, 1) returned error: %v", err)
		}

		pdf0, err := normal.PDF(0)
		if err != nil {
			t.Fatalf("PDF(0) returned error: %v", err)
		}
		evalFloats(t, pdf0, 0.3989422804014327)

		pdf1, err := normal.PDF(1)
		if err != nil {
			t.Fatalf("PDF(1) returned error: %v", err)
		}
		evalFloats(t, pdf1, 0.24197072451914337)

		cdf0, err := normal.CDF(0)
		if err != nil {
			t.Fatalf("CDF(0) returned error: %v", err)
		}
		evalFloats(t, cdf0, 0.5)

		cdf1, err := normal.CDF(1)
		if err != nil {
			t.Fatalf("CDF(1) returned error: %v", err)
		}
		evalFloats(t, cdf1, 0.8413447460685429)

		cdfNeg1, err := normal.CDF(-1)
		if err != nil {
			t.Fatalf("CDF(-1) returned error: %v", err)
		}
		evalFloats(t, cdfNeg1, 0.15865525393145707)

		evalFloats(t, normal.Avg(), 0)
		evalFloats(t, normal.Variance(), 1)
		evalFloats(t, normal.StdDev(), 1)
	})

	t.Run("N(2,3) pinned point values", func(t *testing.T) {
		normal, err := NewNormal(2, 3)
		if err != nil {
			t.Fatalf("NewNormal(2, 3) returned error: %v", err)
		}

		pdf2, err := normal.PDF(2)
		if err != nil {
			t.Fatalf("PDF(2) returned error: %v", err)
		}
		evalFloats(t, pdf2, 0.1329807601338109)

		cdf2, err := normal.CDF(2)
		if err != nil {
			t.Fatalf("CDF(2) returned error: %v", err)
		}
		evalFloats(t, cdf2, 0.5)

		evalFloats(t, normal.Avg(), 2)
		evalFloats(t, normal.Variance(), 9)
		evalFloats(t, normal.StdDev(), 3)
	})

	t.Run("constructor rejects non-positive standard deviation", func(t *testing.T) {
		if got, err := NewNormal(1, 0); err == nil || got != nil {
			t.Errorf("NewNormal(1, 0) = (%v, %v), want (nil, non-nil error)", got, err)
		}
		if got, err := NewNormal(1, -2); err == nil || got != nil {
			t.Errorf("NewNormal(1, -2) = (%v, %v), want (nil, non-nil error)", got, err)
		}
	})
}
