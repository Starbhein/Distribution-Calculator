package sim

import "testing"

// BenchmarkFillGeometricSmallP measures the O(1) inverse-CDF fill at
// p=0.001 over a 1e6-sample buffer (design §3.2). Spec §5 acceptance:
// ≥10× faster than the pre-change O(1/p) iterative baseline
// (~2.22 s/op for the same workload).
func BenchmarkFillGeometricSmallP(b *testing.B) {
	const p = 0.001
	buffer := make([]float64, 1_000_000)
	eng := NewSimulatorEngine(42, 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := eng.FillGeometric(buffer, p); err != nil {
			b.Fatal(err)
		}
	}
}
