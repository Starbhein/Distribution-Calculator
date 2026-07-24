package sim

import (
	"math"
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
