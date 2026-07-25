package registry

import (
	"testing"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
	"github.com/Starbhein/DistCalc/internal/core/distributions/sim"
)

// Tests for the registry skeleton (design §1.3): Spec, ByName, and the
// Sampler interface, exercised through the four continuous entries
// (Normal, Exponential (λ), Exponential (β), Uniforme continua).
// DisplayName and ParamLabels strings are byte-identical to the current UI
// menu/form strings (mainmenu.go:90, formmodel.go:146 switch).

func evalFloats(t testing.TB, got, want float64) {
	t.Helper()
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff > 1e-12 {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestByNameContinuous(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		wantID      ID
		wantLabels  []string
	}{
		{"normal", "Normal", IDNormal, []string{"Media (μ)", "Desviación estándar (σ)"}},
		{"exponential lambda", "Exponencial (λ)", IDExponentialLambda, []string{"Lambda (λ)"}},
		{"exponential beta", "Exponencial (β)", IDExponentialBeta, []string{"Beta (β)"}},
		{"uniform", "Uniforme continua", IDUniform, []string{"Límite inferior (a)", "Límite superior (b)"}},
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
			if spec.Discrete {
				t.Errorf("Discrete: got true, want false for %q", tc.displayName)
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

func TestByNameUnknown(t *testing.T) {
	if spec, ok := ByName("No Existe"); ok {
		t.Fatalf("ByName(\"No Existe\") returned ok=true with spec %+v", spec)
	}
}

func TestConstructContinuous(t *testing.T) {
	t.Run("normal delegates to struct", func(t *testing.T) {
		spec, ok := ByName("Normal")
		if !ok {
			t.Fatal("ByName(\"Normal\") returned ok=false")
		}
		d, err := spec.Construct([]float64{10, 2})
		if err != nil {
			t.Fatalf("Construct returned error: %v", err)
		}
		cd, ok := d.(distributions.ContinuousDistribution)
		if !ok {
			t.Fatalf("Construct did not return a ContinuousDistribution: %T", d)
		}
		want, err := distributions.NewNormal(10, 2)
		if err != nil {
			t.Fatalf("NewNormal returned error: %v", err)
		}
		evalFloats(t, cd.Avg(), want.Avg())
		evalFloats(t, cd.Variance(), want.Variance())
		evalFloats(t, cd.StdDev(), want.StdDev())
		pdfGot, err := cd.PDF(10)
		if err != nil {
			t.Fatalf("PDF returned error: %v", err)
		}
		pdfWant, _ := want.PDF(10)
		evalFloats(t, pdfGot, pdfWant)
	})

	t.Run("exponential lambda delegates to struct", func(t *testing.T) {
		spec, _ := ByName("Exponencial (λ)")
		d, err := spec.Construct([]float64{2})
		if err != nil {
			t.Fatalf("Construct returned error: %v", err)
		}
		want, _ := distributions.NewExponentialLambda(2)
		evalFloats(t, d.Avg(), want.Avg())
		evalFloats(t, d.Variance(), want.Variance())
	})

	t.Run("exponential beta delegates to struct", func(t *testing.T) {
		spec, _ := ByName("Exponencial (β)")
		d, err := spec.Construct([]float64{2})
		if err != nil {
			t.Fatalf("Construct returned error: %v", err)
		}
		want, _ := distributions.NewExponentialBeta(2)
		evalFloats(t, d.Avg(), want.Avg())
		evalFloats(t, d.Variance(), want.Variance())
	})

	t.Run("uniform delegates to struct", func(t *testing.T) {
		spec, _ := ByName("Uniforme continua")
		d, err := spec.Construct([]float64{2, 4})
		if err != nil {
			t.Fatalf("Construct returned error: %v", err)
		}
		want, _ := distributions.NewUniform(2, 4)
		evalFloats(t, d.Avg(), want.Avg())
		evalFloats(t, d.Variance(), want.Variance())
		evalFloats(t, d.StdDev(), want.StdDev())
	})
}

func TestConstructInvalidParams(t *testing.T) {
	cases := []struct {
		name         string
		displayName  string
		params       []float64
		wantIndex    int
	}{
		{"normal sigma zero", "Normal", []float64{10, 0}, 1},
		{"normal sigma negative", "Normal", []float64{10, -2}, 1},
		{"exponential lambda negative", "Exponencial (λ)", []float64{-1}, 0},
		{"exponential lambda zero", "Exponencial (λ)", []float64{0}, 0},
		{"exponential beta negative", "Exponencial (β)", []float64{-1}, 0},
		{"uniform a >= b", "Uniforme continua", []float64{4, 2}, 0},
		{"uniform a == b", "Uniforme continua", []float64{3, 3}, 0},
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

func TestValidateAcceptsCurrentValidParams(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		params      []float64
	}{
		{"normal", "Normal", []float64{10, 2}},
		{"exponential lambda", "Exponencial (λ)", []float64{2}},
		{"exponential beta", "Exponencial (β)", []float64{2}},
		{"uniform", "Uniforme continua", []float64{2, 4}},
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

func TestSamplerContinuous(t *testing.T) {
	engine := sim.NewSimulatorEngine(42, 42)
	buffer := make([]float64, 1000)

	cases := []struct {
		name        string
		displayName string
		params      []float64
		wantAvg     float64
		tolerance   float64
	}{
		{"normal", "Normal", []float64{10, 2}, 10, 0.5},
		{"exponential lambda", "Exponencial (λ)", []float64{2}, 0.5, 0.15},
		{"exponential beta", "Exponencial (β)", []float64{2}, 2, 0.5},
		{"uniform", "Uniforme continua", []float64{2, 4}, 3, 0.3},
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

func TestFormatParamsContinuous(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		params      []float64
		want        string
	}{
		{"normal", "Normal", []float64{10, 2}, "μ=10.0000, σ=2.0000"},
		{"exponential lambda", "Exponencial (λ)", []float64{2}, "λ=2.0000"},
		{"exponential beta", "Exponencial (β)", []float64{2}, "β=2.0000"},
		{"uniform", "Uniforme continua", []float64{2, 4}, "a=2.0000, b=4.0000"},
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
