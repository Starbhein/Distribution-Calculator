package export

import (
	"math"
	"testing"
)

// naiveEmpiricalStats is the pre-PR3b two-pass reference formula, kept here
// ONLY as a test oracle: the production path must delegate to
// stats.AnalyzeBuffer (Welford) and match this within 1e-12 (design §2.5).
func naiveEmpiricalStats(data []float64) (mean, stddev float64) {
	if len(data) == 0 {
		return 0, 0
	}
	var sum float64
	for _, v := range data {
		sum += v
	}
	mean = sum / float64(len(data))
	var variance float64
	for _, v := range data {
		diff := v - mean
		variance += diff * diff
	}
	if len(data) == 1 {
		variance = 0
	} else {
		variance /= float64(len(data) - 1)
	}
	return mean, math.Sqrt(variance)
}

// TestEmpiricalStatsSingleElement pins the n=1 contract (spec §4): the old
// naive code divided by len(data)-1 == 0 at plot.go:203-205 before the guard
// ran; the Welford delegation is well-defined by construction — variance 0
// per stats.AnalyzeBuffer semantics, no NaN.
func TestEmpiricalStatsSingleElement(t *testing.T) {
	mean, stddev := empiricalStats([]float64{7.5})
	if math.IsNaN(mean) || math.IsNaN(stddev) || math.IsInf(mean, 0) || math.IsInf(stddev, 0) {
		t.Fatalf("n=1 produced non-finite stats: mean=%v stddev=%v", mean, stddev)
	}
	if mean != 7.5 {
		t.Errorf("mean = %v, want 7.5", mean)
	}
	if stddev != 0 {
		t.Errorf("stddev = %v, want 0 (AnalyzeBuffer semantics)", stddev)
	}
}

// TestEmpiricalStatsMatchesNaive proves the Welford delegation reproduces
// the deleted naive two-pass outputs within 1e-12 for n>=2 (design §2.5).
func TestEmpiricalStatsMatchesNaive(t *testing.T) {
	const tolerance = 1e-12
	tests := []struct {
		name string
		data []float64
	}{
		{"two elements", []float64{1, 3}},
		{"typical empirical sample", []float64{2.5, 3.1, 2.9, 3.3, 2.7, 3.0, 2.8}},
		{"constant buffer (zero variance)", []float64{4, 4, 4, 4}},
		{"negative and positive mix", []float64{-2.5, 0, 2.5, -1.5, 1.5}},
		{"large magnitudes", []float64{1e6, 1e6 + 1, 1e6 + 2, 1e6 + 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantMean, wantStddev := naiveEmpiricalStats(tt.data)
			gotMean, gotStddev := empiricalStats(tt.data)
			if math.Abs(gotMean-wantMean) > tolerance {
				t.Errorf("mean = %v, want %v (diff %v)", gotMean, wantMean, gotMean-wantMean)
			}
			if math.Abs(gotStddev-wantStddev) > tolerance {
				t.Errorf("stddev = %v, want %v (diff %v)", gotStddev, wantStddev, gotStddev-wantStddev)
			}
		})
	}
}

// TestEmpiricalStatsEmpty pins the empty-buffer behavior: (0, 0), unchanged
// from the deleted naive implementation.
func TestEmpiricalStatsEmpty(t *testing.T) {
	mean, stddev := empiricalStats(nil)
	if mean != 0 || stddev != 0 {
		t.Errorf("empty buffer: got (%v, %v), want (0, 0)", mean, stddev)
	}
}
