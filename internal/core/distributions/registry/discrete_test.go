package registry

import (
	"testing"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
	"github.com/Starbhein/DistCalc/internal/core/distributions/sim"
)

// Tests for the five discrete spec entries (PR4a-2 scope, spec §3, design
// §1.3/§4): Bernoulli, Binomial, Geometric, Poisson, Hypergeometric.
// DisplayName strings are byte-identical to the UI menu (mainmenu.go:90);
// ParamLabels are byte-identical to the formmodel.go:146 switch; Validate
// rules mirror ui.ValidateParams including the hypergeometric cross-param
// rules (M<=N, n<=N) that backstop the latent HypergeometricCDFTable panic
// (design §7).

func TestByNameDiscrete(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		wantID      ID
		wantLabels  []string
	}{
		{"bernoulli", "Bernoulli", IDBernoulli, []string{"Probabilidad (p)"}},
		{"binomial", "Binomial", IDBinomial, []string{"Probabilidad (p)", "Ensayos (N)"}},
		{"geometric", "Geométrica", IDGeometric, []string{"Probabilidad (p)"}},
		{"poisson", "Poisson", IDPoisson, []string{"Lambda (λ)"}},
		{"hypergeometric", "Hypergeométrica", IDHypergeometric, []string{"Tamaño poblacional (N)", "Número de exitos (M)", "Tamaño de muestra (n)"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := ByName(tc.displayName)
			if !ok {
				t.Fatalf("ByName(%q) returned ok=false", tc.displayName)
			}
			if spec.ID != tc.wantID {
				t.Errorf("ID: got %q, want %q", spec.ID, tc.wantID)
			}
			if spec.DisplayName != tc.displayName {
				t.Errorf("DisplayName: got %q, want %q", spec.DisplayName, tc.displayName)
			}
			if !spec.Discrete {
				t.Errorf("Discrete: got false, want true for %q", tc.displayName)
			}
			if len(spec.ParamLabels) != len(tc.wantLabels) {
				t.Fatalf("ParamLabels: got %v, want %v", spec.ParamLabels, tc.wantLabels)
			}
			for i, label := range tc.wantLabels {
				if spec.ParamLabels[i] != label {
					t.Errorf("ParamLabels[%d]: got %q, want %q", i, spec.ParamLabels[i], label)
				}
			}
		})
	}
}

// TestAllNineMenuEntriesRegistered proves the registry now covers every
// hardcoded menu entry (PR4a-2 exit evidence): 4 continuous (PR4a-1) + 5
// discrete (this batch).
func TestAllNineMenuEntriesRegistered(t *testing.T) {
	menuNames := []string{
		"Binomial", "Poisson", "Hypergeométrica", "Normal", "Exponencial (λ)",
		"Exponencial (β)", "Bernoulli", "Geométrica", "Uniforme continua",
	}
	for _, name := range menuNames {
		if _, ok := ByName(name); !ok {
			t.Errorf("ByName(%q) returned ok=false; menu entry not registered", name)
		}
	}
}

func TestConstructDiscrete(t *testing.T) {
	t.Run("bernoulli delegates to struct", func(t *testing.T) {
		spec, ok := ByName("Bernoulli")
		if !ok {
			t.Fatal("ByName(\"Bernoulli\") returned ok=false")
		}
		d, err := spec.Construct([]float64{0.3})
		if err != nil {
			t.Fatalf("Construct returned error: %v", err)
		}
		dd, ok := d.(distributions.DiscreteDistribution)
		if !ok {
			t.Fatalf("Construct did not return a DiscreteDistribution: %T", d)
		}
		want, err := distributions.NewBernoulli(0.3)
		if err != nil {
			t.Fatalf("NewBernoulli returned error: %v", err)
		}
		evalFloats(t, dd.Avg(), want.Avg())
		evalFloats(t, dd.Variance(), want.Variance())
		evalFloats(t, dd.StdDev(), want.StdDev())
	})

	t.Run("binomial delegates to struct", func(t *testing.T) {
		spec, _ := ByName("Binomial")
		// UI order is (p, n); constructor order is (n, p).
		d, err := spec.Construct([]float64{0.5, 10})
		if err != nil {
			t.Fatalf("Construct returned error: %v", err)
		}
		dd, ok := d.(distributions.DiscreteDistribution)
		if !ok {
			t.Fatalf("Construct did not return a DiscreteDistribution: %T", d)
		}
		want, err := distributions.NewBinomial(10, 0.5)
		if err != nil {
			t.Fatalf("NewBinomial returned error: %v", err)
		}
		evalFloats(t, dd.Avg(), want.Avg())
		evalFloats(t, dd.Variance(), want.Variance())
		pmfGot, err := dd.PMF(5)
		if err != nil {
			t.Fatalf("PMF returned error: %v", err)
		}
		pmfWant, _ := want.PMF(5)
		evalFloats(t, pmfGot, pmfWant)
	})

	t.Run("geometric delegates to struct", func(t *testing.T) {
		spec, _ := ByName("Geométrica")
		d, err := spec.Construct([]float64{0.25})
		if err != nil {
			t.Fatalf("Construct returned error: %v", err)
		}
		dd, ok := d.(distributions.DiscreteDistribution)
		if !ok {
			t.Fatalf("Construct did not return a DiscreteDistribution: %T", d)
		}
		want, err := distributions.NewGeometric(0.25)
		if err != nil {
			t.Fatalf("NewGeometric returned error: %v", err)
		}
		evalFloats(t, dd.Avg(), want.Avg())
		evalFloats(t, dd.Variance(), want.Variance())
	})

	t.Run("poisson delegates to struct", func(t *testing.T) {
		spec, _ := ByName("Poisson")
		d, err := spec.Construct([]float64{4})
		if err != nil {
			t.Fatalf("Construct returned error: %v", err)
		}
		dd, ok := d.(distributions.DiscreteDistribution)
		if !ok {
			t.Fatalf("Construct did not return a DiscreteDistribution: %T", d)
		}
		want, err := distributions.NewPoisson(4)
		if err != nil {
			t.Fatalf("NewPoisson returned error: %v", err)
		}
		evalFloats(t, dd.Avg(), want.Avg())
		evalFloats(t, dd.Variance(), want.Variance())
		pmfGot, err := dd.PMF(3)
		if err != nil {
			t.Fatalf("PMF returned error: %v", err)
		}
		pmfWant, _ := want.PMF(3)
		evalFloats(t, pmfGot, pmfWant)
	})

	t.Run("hypergeometric delegates to struct with named binding", func(t *testing.T) {
		spec, _ := ByName("Hypergeométrica")
		// UI order (N, M, n) = (12, 3, 4): Avg = n*M/N = 1 (spec §7 pin).
		d, err := spec.Construct([]float64{12, 3, 4})
		if err != nil {
			t.Fatalf("Construct returned error: %v", err)
		}
		dd, ok := d.(distributions.DiscreteDistribution)
		if !ok {
			t.Fatalf("Construct did not return a DiscreteDistribution: %T", d)
		}
		want, err := distributions.NewHypergeometric(3, 12, 4)
		if err != nil {
			t.Fatalf("NewHypergeometric returned error: %v", err)
		}
		evalFloats(t, dd.Avg(), 1.0)
		evalFloats(t, dd.Avg(), want.Avg())
		evalFloats(t, dd.Variance(), want.Variance())
		// PMF(0) = 14/55 (spec §7 pin) — any parameter permutation changes
		// this value, so the assertion fails loudly on a swap.
		pmfGot, err := dd.PMF(0)
		if err != nil {
			t.Fatalf("PMF returned error: %v", err)
		}
		evalFloats(t, pmfGot, 14.0/55.0)
	})
}

func TestConstructDiscreteInvalidParams(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		params      []float64
		wantIndex   int
	}{
		{"bernoulli p zero", "Bernoulli", []float64{0}, 0},
		{"bernoulli p above one", "Bernoulli", []float64{1.5}, 0},
		{"binomial p zero", "Binomial", []float64{0, 10}, 0},
		{"binomial p above one", "Binomial", []float64{1.5, 10}, 0},
		{"binomial n zero", "Binomial", []float64{0.5, 0}, 1},
		{"geometric p zero", "Geométrica", []float64{0}, 0},
		{"geometric p above one", "Geométrica", []float64{2}, 0},
		{"poisson lambda zero", "Poisson", []float64{0}, 0},
		{"poisson lambda negative", "Poisson", []float64{-1}, 0},
		{"hypergeometric arity", "Hypergeométrica", []float64{12, 3}, -1},
		{"hypergeometric N zero", "Hypergeométrica", []float64{0, 3, 4}, 0},
		{"hypergeometric M zero", "Hypergeométrica", []float64{12, 0, 4}, 1},
		{"hypergeometric n zero", "Hypergeométrica", []float64{12, 3, 0}, 2},
		// Cross-param backstop (design §7): rejects combos that would reach
		// the latent HypergeometricCDFTable startK>maxK panic.
		{"hypergeometric M greater than N", "Hypergeométrica", []float64{12, 13, 4}, 1},
		{"hypergeometric n greater than N", "Hypergeométrica", []float64{12, 3, 13}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := ByName(tc.displayName)
			if !ok {
				t.Fatalf("ByName(%q) returned ok=false", tc.displayName)
			}
			if _, err := spec.Construct(tc.params); err == nil {
				t.Fatalf("Construct(%v) expected error, got nil", tc.params)
			}
			if _, err := spec.NewSampler(tc.params); err == nil {
				t.Fatalf("NewSampler(%v) expected error, got nil", tc.params)
			}
			idx, err := spec.Validate(tc.params)
			if err == nil {
				t.Fatalf("Validate(%v) expected error, got nil", tc.params)
			}
			if idx != tc.wantIndex {
				t.Errorf("Validate paramIndex: got %d, want %d", idx, tc.wantIndex)
			}
		})
	}
}

func TestValidateDiscreteAcceptsCurrentValidParams(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		params      []float64
	}{
		{"bernoulli", "Bernoulli", []float64{0.3}},
		{"bernoulli p one", "Bernoulli", []float64{1}},
		{"binomial", "Binomial", []float64{0.5, 10}},
		{"binomial p one", "Binomial", []float64{1, 10}},
		{"geometric", "Geométrica", []float64{0.25}},
		{"poisson", "Poisson", []float64{4}},
		{"hypergeometric", "Hypergeométrica", []float64{12, 3, 4}},
		{"hypergeometric M equals N", "Hypergeométrica", []float64{12, 12, 4}},
		{"hypergeometric n equals N", "Hypergeométrica", []float64{12, 3, 12}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, _ := ByName(tc.displayName)
			if idx, err := spec.Validate(tc.params); err != nil {
				t.Fatalf("Validate(%v) returned error at index %d: %v", tc.params, idx, err)
			}
		})
	}
}

// TestSamplerDiscrete exercises Prebuild + Fill for the table-backed
// discretes on both gate paths (table and non-table) and for the table-less
// discretes (no-op Prebuild), per design §2.4/§1.5.
func TestSamplerDiscrete(t *testing.T) {
	engine := sim.NewSimulatorEngine(42, 42)
	buffer := make([]float64, 10000)

	cases := []struct {
		name        string
		displayName string
		params      []float64
		wantAvg     float64
		tolerance   float64
	}{
		// UI order (p, n); variance 2.5 <= 9 → CDF-table path.
		{"binomial table path", "Binomial", []float64{0.5, 10}, 5, 0.15},
		// variance 250 > 9 → normal-approximation path (nil table).
		{"binomial normal path", "Binomial", []float64{0.5, 1000}, 500, 2.5},
		// 10 < λ <= 100 → CDF-table path.
		{"poisson table path", "Poisson", []float64{50}, 50, 1.0},
		// λ <= 10 → iterative path (nil table).
		{"poisson iterative path", "Poisson", []float64{5}, 5, 0.35},
		// variance ~0.545 <= 9 → CDF-table path; Avg = n*M/N = 1 (spec §7).
		{"hypergeometric table path", "Hypergeométrica", []float64{12, 3, 4}, 1, 0.1},
		// Table-less: no-op Prebuild.
		{"bernoulli", "Bernoulli", []float64{0.3}, 0.3, 0.05},
		{"geometric", "Geométrica", []float64{0.25}, 4, 0.4},
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
			if err := sampler.Fill(engine, buffer); err != nil {
				t.Fatalf("Fill returned error: %v", err)
			}
			var sum float64
			for _, v := range buffer {
				sum += v
			}
			mean := sum / float64(len(buffer))
			diff := mean - tc.wantAvg
			if diff < 0 {
				diff = -diff
			}
			if diff > tc.tolerance {
				t.Errorf("sample mean: got %v, want %v ± %v", mean, tc.wantAvg, tc.tolerance)
			}
		})
	}
}

func TestFormatParamsDiscrete(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		params      []float64
		want        string
	}{
		{"binomial", "Binomial", []float64{0.5, 10}, "p=0.5000, n=10"},
		{"poisson", "Poisson", []float64{4}, "λ=4.0000"},
		{"hypergeometric", "Hypergeométrica", []float64{12, 3, 4}, "N=12, M=3, n=4"},
		{"bernoulli", "Bernoulli", []float64{0.3}, "p=0.3000"},
		{"geometric", "Geométrica", []float64{0.25}, "p=0.2500"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := ByName(tc.displayName)
			if !ok {
				t.Fatalf("ByName(%q) returned ok=false", tc.displayName)
			}
			if got := spec.FormatParams(tc.params); got != tc.want {
				t.Errorf("FormatParams(%v): got %q, want %q", tc.params, got, tc.want)
			}
		})
	}
}

// TestBindHypergeometric pins the single-sourced named binding (design §4):
// UI order (N, M, n) maps to named fields exactly once, and every consumer
// (Construct, Sampler) references fields — never indices.
func TestBindHypergeometric(t *testing.T) {
	p, err := bindHypergeometric([]float64{12, 3, 4})
	if err != nil {
		t.Fatalf("bindHypergeometric returned error: %v", err)
	}
	if p.Population != 12 {
		t.Errorf("Population: got %d, want 12 (UI position 0)", p.Population)
	}
	if p.Successes != 3 {
		t.Errorf("Successes: got %d, want 3 (UI position 1)", p.Successes)
	}
	if p.Sample != 4 {
		t.Errorf("Sample: got %d, want 4 (UI position 2)", p.Sample)
	}

	if _, err := bindHypergeometric([]float64{12, 3}); err == nil {
		t.Error("bindHypergeometric with 2 params expected error, got nil")
	}
	if _, err := bindHypergeometric([]float64{12, 13, 4}); err == nil {
		t.Error("bindHypergeometric with M>N expected error, got nil")
	}
}

// TestHypergeometricSamplerBinding proves the sampler maps the named fields
// to the engine order (m, nsample, n) = (Successes, Sample, Population):
// the empirical mean over the asymmetric support must agree with the
// theoretical Avg = n*M/N = 1 (pinned exactly in TestConstructDiscrete);
// any permutation yields a different mean. The tolerance is statistical
// (SE ≈ 0.0074 for variance ~0.545 over 10000 samples ⇒ ±0.05 is ~6σ).
func TestHypergeometricSamplerBinding(t *testing.T) {
	spec, ok := ByName("Hypergeométrica")
	if !ok {
		t.Fatal("ByName(\"Hypergeométrica\") returned ok=false")
	}
	sampler, err := spec.NewSampler([]float64{12, 3, 4})
	if err != nil {
		t.Fatalf("NewSampler returned error: %v", err)
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
	diff := mean - 1.0
	if diff < 0 {
		diff = -diff
	}
	if diff > 0.05 {
		t.Errorf("sample mean: got %v, want 1 ± 0.05", mean)
	}
}
