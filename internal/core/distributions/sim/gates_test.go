package sim

import (
	"math"
	"testing"
)

// Table-driven gate-boundary tests for the single-sourced table-vs-normal
// gating predicates (design §2.4). Boundaries under test:
//   - Binomial/Hypergeometric: use the CDF table iff variance <= 9.
//   - Poisson: use the CDF table iff 10 < lambda <= 100.
func TestBinomialUsesTable(t *testing.T) {
	tests := []struct {
		name string
		n    int
		p    float64
		want bool
	}{
		{"variance exactly 9 uses table (boundary inclusive)", 36, 0.5, true},
		{"variance above 9 uses normal approximation", 40, 0.5, false},
		{"small variance uses table", 10, 0.1, true},
		{"existing sim test case n=50 p=0.5 (variance 12.5) uses normal", 50, 0.5, false},
		{"degenerate p=0 has zero variance and uses table", 100, 0.0, true},
		{"NaN p keeps base table polarity (NaN > 9 is false)", 100, math.NaN(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BinomialUsesTable(tt.n, tt.p); got != tt.want {
				t.Errorf("BinomialUsesTable(%d, %v) = %v, want %v", tt.n, tt.p, got, tt.want)
			}
		})
	}
}

func TestPoissonUsesTable(t *testing.T) {
	tests := []struct {
		name   string
		lambda float64
		want   bool
	}{
		{"lambda exactly 10 does not use table (lower boundary exclusive)", 10, false},
		{"lambda just above 10 uses table", 10.5, true},
		{"lambda 100 uses table (upper boundary inclusive)", 100, true},
		{"lambda just above 100 uses normal approximation", 100.5, false},
		{"small lambda uses iterative method", 5, false},
		{"existing sim test case lambda=15 uses table", 15, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PoissonUsesTable(tt.lambda); got != tt.want {
				t.Errorf("PoissonUsesTable(%v) = %v, want %v", tt.lambda, got, tt.want)
			}
		})
	}
}

func TestHypergeometricUsesTable(t *testing.T) {
	tests := []struct {
		name          string
		m, nsample, n float64
		want          bool
	}{
		{"existing sim test case m=20 nsample=10 n=50 (variance ~1.959) uses table", 20, 10, 50, true},
		{"variance just below 9 uses table (N=144 M=72 K=60, variance ~8.811)", 72, 60, 144, true},
		{"variance just above 9 uses normal (N=144 M=72 K=68, variance ~9.035)", 72, 68, 144, false},
		{"large variance uses normal (N=500 M=250 K=100, variance ~20.04)", 250, 100, 500, false},
		{"degenerate N=1 support has NaN variance and keeps base table polarity", 1, 1, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HypergeometricUsesTable(tt.m, tt.nsample, tt.n); got != tt.want {
				t.Errorf("HypergeometricUsesTable(%v, %v, %v) = %v, want %v",
					tt.m, tt.nsample, tt.n, got, tt.want)
			}
		})
	}
}
