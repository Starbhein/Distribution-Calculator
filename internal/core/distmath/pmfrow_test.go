package distmath

import (
	"math"
	"testing"
)

// Tests for the PMF row adapters (PR4c scope, design §3.3): one O(range)
// pass per render for binomial and poisson, mirroring HypergeometricPMFRow.
// Every value must agree with the pointwise closed form within the pinned
// 1e-12 tolerance so rendered output is unchanged (spec §5).

func TestBinomialPMFRowPinnedValues(t *testing.T) {
	// Binomial(n=10, p=0.1): PMF(2) is pinned at the struct layer; the row
	// must reproduce it within tolerance.
	row := BinomialPMFRow(10, 0.1)
	if len(row) != 11 {
		t.Fatalf("BinomialPMFRow(10, 0.1) returned %d entries, want 11", len(row))
	}
	want := math.Exp(logBinomialPMF(10, 0.1, 2))
	evalFloats(t, row[2], want)
}

func TestBinomialPMFRowMatchesClosedForm(t *testing.T) {
	tests := []struct {
		name string
		n    int
		p    float64
	}{
		{"n=1 p=0.5 (edge support)", 1, 0.5},
		{"n=10 p=0.5 (symmetric)", 10, 0.5},
		{"n=40 p=0.25 (asymmetric)", 40, 0.25},
		{"n=200 p=0.9 (right-heavy)", 200, 0.9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := BinomialPMFRow(tt.n, tt.p)
			if len(row) != tt.n+1 {
				t.Fatalf("row length = %d, want %d", len(row), tt.n+1)
			}
			sum := 0.0
			for k := 0; k <= tt.n; k++ {
				evalFloats(t, row[k], math.Exp(logBinomialPMF(tt.n, tt.p, k)))
				sum += row[k]
			}
			evalFloats(t, sum, 1.0)
		})
	}
}

func TestPoissonPMFRowPinnedValues(t *testing.T) {
	// Poisson(λ=4): PMF(3) = 0.19536681481316 is pinned at the struct layer
	// (spec §4); the row must reproduce it within tolerance.
	row := PoissonPMFRow(4)
	if len(row) < 4 {
		t.Fatalf("PoissonPMFRow(4) returned %d entries, want at least 4", len(row))
	}
	evalFloats(t, row[3], 0.19536681481316)
}

func TestPoissonPMFRowMatchesClosedForm(t *testing.T) {
	tests := []struct {
		name   string
		lambda float64
	}{
		{"lambda=0.5 (tiny, short row)", 0.5},
		{"lambda=4 (pinned)", 4},
		{"lambda=25 (wider support)", 25},
		{"lambda=100 (table gate upper bound)", 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := PoissonPMFRow(tt.lambda)
			if len(row) < 2 {
				t.Fatalf("row length = %d, want at least 2", len(row))
			}
			sum := 0.0
			for k, v := range row {
				evalFloats(t, v, math.Exp(logPoissonPMF(tt.lambda, k)))
				sum += v
			}
			// The row covers lambda + 4*sigma + 1: >=99.98% of mass even
			// for the tiniest lambda in the table (0.5).
			if sum < 0.999 {
				t.Errorf("row mass = %v, want > 0.999", sum)
			}
		})
	}
}

// TestPoissonPMFRowExtentMatchesCDFTable pins the row extent to the CDF
// table's kMax rule so both consumers see the same support.
func TestPoissonPMFRowExtentMatchesCDFTable(t *testing.T) {
	for _, lambda := range []float64{0.5, 4, 25, 100} {
		row := PoissonPMFRow(lambda)
		table := PoissonCDFTable(lambda)
		if len(row) != len(table) {
			t.Errorf("lambda=%v: row length %d != CDF table length %d", lambda, len(row), len(table))
		}
	}
}
