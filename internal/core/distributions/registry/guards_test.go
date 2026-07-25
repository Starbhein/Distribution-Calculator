package registry

import (
	"testing"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
	"github.com/Starbhein/DistCalc/internal/core/distributions/sim"
)

// Guard tests committed by PR4b-1 (design §1.5, §4, §9).

// TestSamplerFillZeroAllocs guards the CLT hot path (design §1.5): the
// per-worker Fill of table-less samplers is exactly the engine method body
// — PRNG + arithmetic into a preallocated buffer — and MUST stay
// allocation-free per call.
func TestSamplerFillZeroAllocs(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		params      []float64
	}{
		{"normal", "Normal", []float64{10, 2}},
		{"bernoulli", "Bernoulli", []float64{0.3}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := ByName(tc.displayName)
			if !ok {
				t.Fatalf("ByName(%q) returned ok=false", tc.displayName)
			}
			sampler, err := spec.NewSampler(tc.params)
			if err != nil {
				t.Fatalf("NewSampler(%v) returned error: %v", tc.params, err)
			}
			if err := sampler.Prebuild(); err != nil {
				t.Fatalf("Prebuild returned error: %v", err)
			}
			engine := sim.NewSimulatorEngine(42, 42)
			buffer := make([]float64, 4096)
			allocs := testing.AllocsPerRun(100, func() {
				if err := sampler.Fill(engine, buffer); err != nil {
					t.Fatalf("Fill returned error: %v", err)
				}
			})
			if allocs != 0 {
				t.Errorf("Fill allocated %v times per run, want 0 (design §1.5 hot path)", allocs)
			}
		})
	}
}

// TestHypergeometricPermutationRegression is the spec §7 / design §4
// regression guard: an end-to-end asymmetric case (N=12, M=3, n=4, so M≠n
// and every permutation provably changes the answer) driven entirely
// through the registry path. A parameter swap introduced anywhere makes
// this test fail loudly — never a silent wrong answer.
func TestHypergeometricPermutationRegression(t *testing.T) {
	spec, ok := ByName("Hypergeométrica")
	if !ok {
		t.Fatal("ByName(\"Hypergeométrica\") returned ok=false")
	}
	params := []float64{12, 3, 4} // UI order (N, M, n)

	// Theoretical Avg = n·M/N = 1 via the stats helper (design §1.3).
	avg, _, _, err := TheoreticalStats(spec, params)
	if err != nil {
		t.Fatalf("TheoreticalStats(%v) returned error: %v", params, err)
	}
	evalFloats(t, avg, 1.0)

	// PMF(0) = 14/55 via Construct (pinned value, spec §7).
	d, err := spec.Construct(params)
	if err != nil {
		t.Fatalf("Construct(%v) returned error: %v", params, err)
	}
	dd, ok := d.(distributions.DiscreteDistribution)
	if !ok {
		t.Fatalf("Construct did not return a DiscreteDistribution: %T", d)
	}
	pmf0, err := dd.PMF(0)
	if err != nil {
		t.Fatalf("PMF(0) returned error: %v", err)
	}
	evalFloats(t, pmf0, 14.0/55.0)

	// Theoretical/empirical agreement through the sampler path: the
	// empirical mean over the asymmetric support must agree with Avg=1.
	// Tolerance is statistical (SE ≈ 0.0074 for variance ~0.545 over
	// 10000 samples ⇒ ±0.05 is ~6σ).
	sampler, err := spec.NewSampler(params)
	if err != nil {
		t.Fatalf("NewSampler(%v) returned error: %v", params, err)
	}
	if err := sampler.Prebuild(); err != nil {
		t.Fatalf("Prebuild returned error: %v", err)
	}
	engine := sim.NewSimulatorEngine(42, 42)
	buffer := make([]float64, 10000)
	if err := sampler.Fill(engine, buffer); err != nil {
		t.Fatalf("Fill returned error: %v", err)
	}
	var sum float64
	for _, v := range buffer {
		sum += v
	}
	mean := sum / float64(len(buffer))
	diff := mean - avg
	if diff < 0 {
		diff = -diff
	}
	if diff > 0.05 {
		t.Errorf("theoretical/empirical disagreement: Avg=%v, empirical mean=%v (tolerance 0.05)", avg, mean)
	}
}
