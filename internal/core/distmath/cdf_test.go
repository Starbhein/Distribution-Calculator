package distmath

import (
	"math"
	"testing"
)

// TestEpsilonSignificantValueSingleHome pins the design §2.3 decision:
// 1e-18 is THE single truncation epsilon for every recurrence core in this
// package (the struct layer's 1e-15 and the sim layer's 1e-18 converge here;
// the added tail mass to struct CDFs is <=1e-15 relative, two orders below
// the 1e-12 pinned-test tolerance).
func TestEpsilonSignificantValueSingleHome(t *testing.T) {
	if EpsilonSignificantValue != 1e-18 {
		t.Errorf("EpsilonSignificantValue = %v, want exactly 1e-18 (design §2.3)", EpsilonSignificantValue)
	}
}

// TestBinomialCDFPinnedValues pins spec §4 acceptance: Binomial.CDF(N) = 1
// forced exactly (design §2.2), plus the n=1000 benchmark case which must
// stay finite and correct (the pre-PR3a struct anchor N*int(p) collapsed to
// 0 and overflowed the walk to +Inf on this input).
func TestBinomialCDFPinnedValues(t *testing.T) {
	t.Run("CDF(N) is forced to exactly 1", func(t *testing.T) {
		if got := BinomialCDF(10, 0.5, 10); got != 1.0 {
			t.Errorf("BinomialCDF(10, 0.5, 10) = %v, want exactly 1.0", got)
		}
	})
	t.Run("CDF below 0 is 0", func(t *testing.T) {
		if got := BinomialCDF(10, 0.5, -1); got != 0 {
			t.Errorf("BinomialCDF(10, 0.5, -1) = %v, want 0", got)
		}
	})
	t.Run("benchmark case n=1000 p=0.9 CDF(999) is 1 within tolerance", func(t *testing.T) {
		// True value: 1 - 0.9^1000 ≈ 1 - 1.7e-46 → 1 at float64 precision.
		got := BinomialCDF(1000, 0.9, 999)
		evalFloats(t, got, 1)
	})
	t.Run("CDF mid-range matches PMF partial sums", func(t *testing.T) {
		// n=10, p=0.1: CDF(2) = PMF(0)+PMF(1)+PMF(2) with the pinned
		// PMF(2) = 0.1937102445 from binomial_test.go.
		const wantPMF2 = .1937102445
		pmf0 := math.Pow(0.9, 10)
		pmf1 := 10 * 0.1 * math.Pow(0.9, 9)
		evalFloats(t, BinomialCDF(10, 0.1, 2), pmf0+pmf1+wantPMF2)
	})
}

// TestPoissonCDFPinnedValues pins spec §4 acceptance via the pre-existing
// pinned struct values: PMF(3; λ=4) = 0.19536681481316 and
// CDF(3; λ=4) = 0.433470120366708933617 (poisson_test.go, tolerance 1e-12).
// The kernel is tied to the pinned PMF through CDF(3)-CDF(2) = PMF(3).
func TestPoissonCDFPinnedValues(t *testing.T) {
	const (
		lambda  = 4.0
		wantPMF = 0.19536681481316
		wantCDF = 0.433470120366708933617
	)
	t.Run("CDF(3) matches pinned struct value", func(t *testing.T) {
		evalFloats(t, PoissonCDF(lambda, 3), wantCDF)
	})
	t.Run("CDF(3)-CDF(2) equals pinned PMF(3)", func(t *testing.T) {
		got := PoissonCDF(lambda, 3) - PoissonCDF(lambda, 2)
		evalFloats(t, got, wantPMF)
	})
	t.Run("CDF below 0 is 0", func(t *testing.T) {
		if got := PoissonCDF(lambda, -1); got != 0 {
			t.Errorf("PoissonCDF(4, -1) = %v, want 0", got)
		}
	})
	t.Run("CDF far tail approaches 1", func(t *testing.T) {
		evalFloats(t, PoissonCDF(lambda, 30), 1)
	})
}

// TestHypergeometricCDFPinnedValues pins spec §4 acceptance:
// Hypergeometric.CDF(2) = 54/55 for M=3, N=12, n=4 (tolerance 1e-12), plus
// the support-boundary contract shared with HypergeometricPMF.
func TestHypergeometricCDFPinnedValues(t *testing.T) {
	const (
		m       = 3
		nSample = 4
		n       = 12
	)
	t.Run("CDF(2) equals 54/55", func(t *testing.T) {
		const want = float64(54) / float64(55)
		evalFloats(t, HypergeometricCDF(m, nSample, n, 2), want)
	})
	t.Run("CDF(maxK) is exactly 1", func(t *testing.T) {
		if got := HypergeometricCDF(m, nSample, n, 3); got != 1.0 {
			t.Errorf("HypergeometricCDF(3, 4, 12, 3) = %v, want exactly 1.0", got)
		}
	})
	t.Run("CDF below startK is 0", func(t *testing.T) {
		// Asymmetric support [3, 10]: CDF(2) must be exactly 0.
		if got := HypergeometricCDF(10, 18, 25, 2); got != 0 {
			t.Errorf("HypergeometricCDF(10, 18, 25, 2) = %v, want 0", got)
		}
	})
	t.Run("CDF over asymmetric support reaches 1", func(t *testing.T) {
		evalFloats(t, HypergeometricCDF(10, 18, 25, 10), 1)
	})
}

// TestBinomialCDFTable verifies the materialized+normalized table adapter:
// full support [0, n], last entry forced to exactly 1, monotone, and
// pointwise/table agreement within the pinned 1e-12 tolerance (same
// recurrence core, design §2.1).
func TestBinomialCDFTable(t *testing.T) {
	const (
		n = 10
		p = 0.5
	)
	table := BinomialCDFTable(n, p)
	if len(table) != n+1 {
		t.Fatalf("table length = %d, want %d", len(table), n+1)
	}
	if table[n] != 1.0 {
		t.Errorf("table[n] = %v, want exactly 1.0", table[n])
	}
	for k := 0; k <= n; k++ {
		if k > 0 && table[k] < table[k-1] {
			t.Errorf("table not monotone at k=%d: %v < %v", k, table[k], table[k-1])
		}
		evalFloats(t, table[k], BinomialCDF(n, p, k))
	}
	// Extreme-p shape: all mass at 0 → table is 1 everywhere.
	low := BinomialCDFTable(10, 0)
	for k, v := range low {
		if v != 1.0 {
			t.Errorf("BinomialCDFTable(10, 0)[%d] = %v, want 1.0", k, v)
		}
	}
}

// TestPoissonCDFTable verifies the truncated, normalized Poisson table. The
// table captures >99.99% of probability mass (kMax = int(λ+4σ)+1), so
// agreement with the exact pointwise kernel is asserted within a documented
// 1e-3 bound — NOT the 1e-12 pinned tolerance (the missing far tail is by
// design; sim sampling is the consumer).
func TestPoissonCDFTable(t *testing.T) {
	const lambda = 4.0
	table := PoissonCDFTable(lambda)
	kMax := int(lambda+4.0*math.Sqrt(lambda)) + 1
	if len(table) != kMax+1 {
		t.Fatalf("table length = %d, want %d", len(table), kMax+1)
	}
	if table[kMax] != 1.0 {
		t.Errorf("table[kMax] = %v, want exactly 1.0", table[kMax])
	}
	for k := 0; k <= kMax; k++ {
		if k > 0 && table[k] < table[k-1] {
			t.Errorf("table not monotone at k=%d: %v < %v", k, table[k], table[k-1])
		}
		if diff := math.Abs(table[k] - PoissonCDF(lambda, k)); diff > 1e-3 {
			t.Errorf("table[%d] = %v vs pointwise %v: diff %v > 1e-3 (truncated-tail bound)",
				k, table[k], PoissonCDF(lambda, k), diff)
		}
	}
}

// TestHypergeometricCDFTable verifies the materialized+normalized
// hypergeometric table adapter over its support [startK, maxK]: last entry
// forced to exactly 1, monotone, pointwise/table agreement within 1e-12,
// and the nsample > n validation error preserved (design §2.1).
func TestHypergeometricCDFTable(t *testing.T) {
	t.Run("pinned support M=3 N=12 n=4", func(t *testing.T) {
		table, startK, maxK, err := HypergeometricCDFTable(3, 4, 12)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if startK != 0 || maxK != 3 {
			t.Fatalf("support = [%d, %d], want [0, 3]", startK, maxK)
		}
		if len(table) != maxK-startK+1 {
			t.Fatalf("table length = %d, want %d", len(table), maxK-startK+1)
		}
		if table[len(table)-1] != 1.0 {
			t.Errorf("table[last] = %v, want exactly 1.0", table[len(table)-1])
		}
		for idx, k := 0, startK; k <= maxK; idx, k = idx+1, k+1 {
			if idx > 0 && table[idx] < table[idx-1] {
				t.Errorf("table not monotone at k=%d: %v < %v", k, table[idx], table[idx-1])
			}
			evalFloats(t, table[idx], HypergeometricCDF(3, 4, 12, k))
		}
		// The pinned CDF(2) = 54/55 must hold through the table as well.
		evalFloats(t, table[2], float64(54)/float64(55))
	})
	t.Run("asymmetric support", func(t *testing.T) {
		table, startK, maxK, err := HypergeometricCDFTable(10, 18, 25)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if startK != 3 || maxK != 10 {
			t.Fatalf("support = [%d, %d], want [3, 10]", startK, maxK)
		}
		if table[len(table)-1] != 1.0 {
			t.Errorf("table[last] = %v, want exactly 1.0", table[len(table)-1])
		}
		for idx, k := 0, startK; k <= maxK; idx, k = idx+1, k+1 {
			evalFloats(t, table[idx], HypergeometricCDF(10, 18, 25, k))
		}
	})
	t.Run("nsample above population errors", func(t *testing.T) {
		if _, _, _, err := HypergeometricCDFTable(3, 13, 12); err == nil {
			t.Error("HypergeometricCDFTable(3, 13, 12): want non-nil error")
		}
	})
}

// TestPointwiseKernelsAllocationFree pins the design §2.1 contract: the
// pointwise adapters materialize nothing — zero heap allocations per call.
// This guards the BenchmarkCDF no-regression acceptance (spec §4).
func TestPointwiseKernelsAllocationFree(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		sink = BinomialCDF(1000, 0.9, 999)
		sink = PoissonCDF(50, 73)
		sink = HypergeometricCDF(50, 40, 500, 20)
	})
	if allocs != 0 {
		t.Errorf("pointwise kernels allocated %v times per run, want 0", allocs)
	}
}
