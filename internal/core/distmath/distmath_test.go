package distmath

import (
	"math"
	"testing"
)

// tolerance matches the pre-existing pinned-test tolerance (binomial_test.go
// epsilonFailure = 1e-12); design §2.2 bounds the recurrence drift at ~1e-15.
const tolerance = 1e-12

func evalFloats(t testing.TB, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Errorf("got %v, want %v (diff %v > %v)", got, want, math.Abs(got-want), tolerance)
	}
}

// logHypergeometricPMFClosedForm is the independent test-side reference: the
// 9-term Lgamma closed form (same expression the old Hypergeometric.PMF used).
// Row/pointwise recurrence results are triangulated against it.
func logHypergeometricPMFClosedForm(m, nSample, n, k int) float64 {
	lg := func(x int) float64 { v, _ := math.Lgamma(float64(x) + 1); return v }
	return lg(m) - lg(k) - lg(m-k) +
		lg(n-m) - lg(nSample-k) - lg(n-m-nSample+k) -
		(lg(n) - lg(nSample) - lg(n-nSample))
}

// TestHypergeometricPMFPinnedValues pins spec §5 / design §3.1 acceptance:
// PMF(0) = 14/55 and ΣPMF = 1 for M=3, N=12, n=4.
func TestHypergeometricPMFPinnedValues(t *testing.T) {
	const (
		m       = 3
		nSample = 4
		n       = 12
	)
	t.Run("PMF(0) equals 14/55", func(t *testing.T) {
		const want = float64(14) / float64(55)
		got := HypergeometricPMF(m, nSample, n, 0)
		evalFloats(t, got, want)
	})
	t.Run("PMF row sums to 1", func(t *testing.T) {
		row, startK, maxK := HypergeometricPMFRow(m, nSample, n)
		if startK != 0 || maxK != 3 {
			t.Fatalf("support = [%d, %d], want [0, 3]", startK, maxK)
		}
		var sum float64
		for _, v := range row {
			sum += v
		}
		evalFloats(t, sum, 1)
	})
	t.Run("pointwise PMF sums to 1 over support", func(t *testing.T) {
		var sum float64
		for k := 0; k <= m; k++ {
			sum += HypergeometricPMF(m, nSample, n, k)
		}
		evalFloats(t, sum, 1)
	})
}

// TestHypergeometricPMFTriangulation checks table-driven (M, nSample, N)
// supports — including asymmetric supports (startK > 0) and edge points
// (k = startK, k = maxK) — against the independent closed form. Bound:
// ~1 ulp per walk step from the mode (design §2.2/§3.1) for the row walk
// itself, plus the closed-form reference's own ~1e-13 Lgamma evaluation
// noise that shows up in the far tail of large supports. All bounds stay
// inside the pre-existing 1e-12 pinned tolerance.
func TestHypergeometricPMFTriangulation(t *testing.T) {
	cases := []struct {
		name             string
		m, nSample, n    int
		wantStart, wantM int     // expected support [startK, maxK]
		relBound         float64 // row/pointwise vs closed-form relative bound
	}{
		{"pinned symmetric-ish support", 3, 4, 12, 0, 3, 1e-13},
		{"single point support", 1, 1, 2, 0, 1, 1e-13},
		{"sample limited support", 5, 3, 20, 0, 3, 1e-13},
		{"success limited support", 4, 10, 20, 0, 4, 1e-13},
		{"asymmetric support startK above zero", 10, 18, 25, 3, 10, 1e-13},
		// Far-tail values (~1e-32 and below) differ from a *separate*
		// 9-Lgamma evaluation by up to ~7e-13 — the bound is the Lgamma
		// reference noise, not the walk (~16 steps ⇒ ~2e-15). Still ≪ 1e-12.
		{"larger realistic support", 50, 40, 500, 0, 40, 1e-12},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			row, startK, maxK := HypergeometricPMFRow(tc.m, tc.nSample, tc.n)
			if startK != tc.wantStart || maxK != tc.wantM {
				t.Fatalf("support = [%d, %d], want [%d, %d]", startK, maxK, tc.wantStart, tc.wantM)
			}
			if len(row) != maxK-startK+1 {
				t.Fatalf("row length = %d, want %d", len(row), maxK-startK+1)
			}

			var sum float64
			for idx, k := 0, startK; k <= maxK; idx, k = idx+1, k+1 {
				ref := math.Exp(logHypergeometricPMFClosedForm(tc.m, tc.nSample, tc.n, k))

				// Row value vs closed form (relative error ~1 ulp/step bound).
				if ref > 0 {
					if rel := math.Abs(row[idx]-ref) / ref; rel > tc.relBound {
						t.Errorf("row[%d] (k=%d) = %v vs closed form %v: relative error %v > %v",
							idx, k, row[idx], ref, rel, tc.relBound)
					}
				}

				// Pointwise PMF vs closed form.
				got := HypergeometricPMF(tc.m, tc.nSample, tc.n, k)
				if ref > 0 {
					if rel := math.Abs(got-ref) / ref; rel > tc.relBound {
						t.Errorf("PMF(k=%d) = %v vs closed form %v: relative error %v > %v",
							k, got, ref, rel, tc.relBound)
					}
				}

				// Pointwise PMF and row must agree (same recurrence core).
				if math.Abs(got-row[idx]) > tolerance {
					t.Errorf("PMF(k=%d) = %v vs row value %v: diff exceeds %v", k, got, row[idx], tolerance)
				}
				sum += row[idx]
			}

			// Every row is a full probability distribution over its support.
			evalFloats(t, sum, 1)
		})
	}
}

// TestHypergeometricPMFOutsideSupport pins the out-of-support contract:
// points outside [startK, maxK] have exactly zero probability.
func TestHypergeometricPMFOutsideSupport(t *testing.T) {
	// M=3, N=12, n=4 → support [0, 3].
	if got := HypergeometricPMF(3, 4, 12, -1); got != 0 {
		t.Errorf("PMF below support = %v, want 0", got)
	}
	if got := HypergeometricPMF(3, 4, 12, 4); got != 0 {
		t.Errorf("PMF above support = %v, want 0", got)
	}
	// Asymmetric support [3, 10]: k=2 must be exactly 0.
	if got := HypergeometricPMF(10, 18, 25, 2); got != 0 {
		t.Errorf("PMF below asymmetric support = %v, want 0", got)
	}
}

// BenchmarkHypergeometricPMFSupportScan measures the O(range) row pass over
// the full support (M=50, N=500, n=40). Spec §5 acceptance: ≥5× faster than
// the pre-change per-k closed-form scan (baseline ~3903 ns/op, 41 points,
// 9 Lgamma calls per point).
func BenchmarkHypergeometricPMFSupportScan(b *testing.B) {
	const (
		m       = 50
		nSample = 40
		n       = 500
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row, _, _ := HypergeometricPMFRow(m, nSample, n)
		var sum float64
		for _, v := range row {
			sum += v
		}
		sink = sum
	}
}

// BenchmarkHypergeometricPMF measures a single point evaluation (record-only,
// no threshold — design §3.1 honesty note: any exact single point needs one
// seed, so it cannot asymptotically beat the old 9-Lgamma closed form).
func BenchmarkHypergeometricPMF(b *testing.B) {
	const (
		m       = 50
		nSample = 40
		n       = 500
	)
	ks := []int{4, 24, 40} // far below mode, near mode (24), max support edge
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = HypergeometricPMF(m, nSample, n, ks[i%len(ks)])
	}
}

var sink float64
