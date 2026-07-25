package sim

import (
	"math"
	"os"
	"testing"

	"github.com/Starbhein/DistCalc/internal/core/stats"
)

// Gate-0 moment contract for the small-p FillGeometric path (spec §2,
// design §6.3). This pins the STATISTICAL behavior of the current O(1/p)
// iterative sampler and must still pass after the O(1) inverse-CDF swap
// in PR2 — it guards the swap without overfitting the PRNG stream.
func TestFillGeometricSmallPMoments(t *testing.T) {
	const (
		p          = 0.001
		bufferSize = 100_000
		seed       = 42
	)

	buffer := make([]float64, bufferSize)
	eng := NewSimulatorEngine(seed, seed)
	if err := eng.FillGeometric(buffer, p); err != nil {
		t.Fatalf("FillGeometric returned error: %v", err)
	}

	result := stats.AnalyzeBuffer(buffer)

	// Theoretical moments: mean = 1/p, variance = (1-p)/p^2.
	// Tolerances from design §6.3: mean within ±1.5% (~4.7σ),
	// variance within ±5% (~5.6σ for geometric kurtosis).
	const wantAvg = 1.0 / p
	const wantVariance = (1.0 - p) / (p * p)

	if avgErr := math.Abs(result.Avg-wantAvg) / wantAvg; avgErr > 0.015 {
		t.Errorf("sample mean %v vs theoretical %v: relative error %v exceeds 1.5%%", result.Avg, wantAvg, avgErr)
	}
	if varErr := math.Abs(result.Variance-wantVariance) / wantVariance; varErr > 0.05 {
		t.Errorf("sample variance %v vs theoretical %v: relative error %v exceeds 5%%", result.Variance, wantVariance, varErr)
	}
}

// TestFillGeometricSmallPMomentsHeavy is the 1e7-sample variant of the Gate-0
// moment contract above. It is gated behind the DISTCALC_HEAVY env var (design:
// "1e7 moment test gate") so it NEVER runs in the default `go test ./...`
// suite nor under `-short`; `make bench-precision` sets the var to run it.
// Env var chosen over testing.Short() (would still run by default) and over a
// build tag (invisible to -list, easy to forget) — the skip stays discoverable
// in test output.
//
// Tolerance derivation (sqrt(n) scaling): the 1e5 test asserts mean ±1.5%
// (~4.7σ) and variance ±5% (~5.6σ for geometric kurtosis). Standard error
// scales as 1/sqrt(n); 1e5 -> 1e7 is 100x samples, so σ shrinks 10x. Keeping
// the SAME σ-margins gives mean ±0.15% (1.5%/10) and variance ±0.5% (5%/10).
// If n ever changes again, rescale tolerances mechanically by sqrt(n).
func TestFillGeometricSmallPMomentsHeavy(t *testing.T) {
	if os.Getenv("DISTCALC_HEAVY") == "" {
		t.Skip("skipping 1e7-sample heavy moment test; set DISTCALC_HEAVY=1 to run it (e.g. DISTCALC_HEAVY=1 go test -run TestFillGeometricSmallPMomentsHeavy ./internal/core/distributions/sim/)")
	}

	const (
		p          = 0.001
		bufferSize = 10_000_000
		seed       = 42
	)

	buffer := make([]float64, bufferSize)
	eng := NewSimulatorEngine(seed, seed)
	if err := eng.FillGeometric(buffer, p); err != nil {
		t.Fatalf("FillGeometric returned error: %v", err)
	}

	result := stats.AnalyzeBuffer(buffer)

	// Theoretical moments: mean = 1/p, variance = (1-p)/p^2.
	// Tolerances derived above: mean within ±0.15% (~4.7σ),
	// variance within ±0.5% (~5.6σ for geometric kurtosis).
	const wantAvg = 1.0 / p
	const wantVariance = (1.0 - p) / (p * p)

	if avgErr := math.Abs(result.Avg-wantAvg) / wantAvg; avgErr > 0.0015 {
		t.Errorf("sample mean %v vs theoretical %v: relative error %v exceeds 0.15%%", result.Avg, wantAvg, avgErr)
	}
	if varErr := math.Abs(result.Variance-wantVariance) / wantVariance; varErr > 0.005 {
		t.Errorf("sample variance %v vs theoretical %v: relative error %v exceeds 0.5%%", result.Variance, wantVariance, varErr)
	}
}
