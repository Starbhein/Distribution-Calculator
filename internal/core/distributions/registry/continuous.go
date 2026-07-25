package registry

import (
	"errors"
	"fmt"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
	"github.com/Starbhein/DistCalc/internal/core/distributions/sim"
)

// Continuous spec entries (PR4a-1 scope). DisplayName and ParamLabels are
// byte-identical to the current UI strings (mainmenu.go:90,
// formmodel.go:146 switch); Validate rules mirror ui.ValidateParams;
// FormatParams output is byte-identical to formatLLNParams /
// formatCLTParams. Discrete entries land in PR4a-2.

// Table-less samplers share a no-op Prebuild (design §1.5).
type continuousSampler struct {
	fill func(engine *sim.SimulatorEngine, buffer []float64) error
}

func (s continuousSampler) Prebuild() error { return nil }

func (s continuousSampler) Fill(engine *sim.SimulatorEngine, buffer []float64) error {
	return s.fill(engine, buffer)
}

// Validate rules mirror ui.ValidateParams exactly (design §7 convergence
// target); the UI validator dies in PR4b when consumers switch to these.
func validateNormal(params []float64) (int, error) {
	if len(params) < 2 {
		return -1, errors.New("faltan parámetros para Normal")
	}
	if params[1] <= 0 {
		return 1, errors.New("σ debe ser mayor que 0")
	}
	return -1, nil
}

func validateExponentialLambda(params []float64) (int, error) {
	if len(params) < 1 {
		return -1, errors.New("faltan parámetros para Exponencial")
	}
	if params[0] <= 0 {
		return 0, errors.New("λ debe ser mayor que 0")
	}
	return -1, nil
}

func validateExponentialBeta(params []float64) (int, error) {
	if len(params) < 1 {
		return -1, errors.New("faltan parámetros para Exponencial (β)")
	}
	if params[0] <= 0 {
		return 0, errors.New("β debe ser mayor que 0")
	}
	return -1, nil
}

func validateUniform(params []float64) (int, error) {
	if len(params) < 2 {
		return -1, errors.New("faltan parámetros para Uniforme continua")
	}
	if params[0] >= params[1] {
		return 0, errors.New("a debe ser menor que b")
	}
	return -1, nil
}

var normalSpec = Spec{
	ID:          IDNormal,
	DisplayName: "Normal",
	Discrete:    false,
	ParamLabels: []string{"Media (μ)", "Desviación estándar (σ)"},
	Validate:    validateNormal,
	Construct: func(params []float64) (distributions.Distribution, error) {
		if idx, err := validateNormal(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		return distributions.NewNormal(params[0], params[1])
	},
	NewSampler: func(params []float64) (Sampler, error) {
		if idx, err := validateNormal(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		avg, stdDev := params[0], params[1]
		return continuousSampler{fill: func(engine *sim.SimulatorEngine, buffer []float64) error {
			return engine.FillNormal(buffer, avg, stdDev)
		}}, nil
	},
	FormatParams: func(params []float64) string {
		if len(params) >= 2 {
			return fmt.Sprintf("μ=%.4f, σ=%.4f", params[0], params[1])
		}
		return ""
	},
}

var exponentialLambdaSpec = Spec{
	ID:          IDExponentialLambda,
	DisplayName: "Exponencial (λ)",
	Discrete:    false,
	ParamLabels: []string{"Lambda (λ)"},
	Validate:    validateExponentialLambda,
	Construct: func(params []float64) (distributions.Distribution, error) {
		if idx, err := validateExponentialLambda(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		return distributions.NewExponentialLambda(params[0])
	},
	NewSampler: func(params []float64) (Sampler, error) {
		if idx, err := validateExponentialLambda(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		lambda := params[0]
		return continuousSampler{fill: func(engine *sim.SimulatorEngine, buffer []float64) error {
			return engine.FillExponential(buffer, lambda)
		}}, nil
	},
	FormatParams: func(params []float64) string {
		if len(params) >= 1 {
			return fmt.Sprintf("λ=%.4f", params[0])
		}
		return ""
	},
}

var exponentialBetaSpec = Spec{
	ID:          IDExponentialBeta,
	DisplayName: "Exponencial (β)",
	Discrete:    false,
	ParamLabels: []string{"Beta (β)"},
	Validate:    validateExponentialBeta,
	Construct: func(params []float64) (distributions.Distribution, error) {
		if idx, err := validateExponentialBeta(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		return distributions.NewExponentialBeta(params[0])
	},
	NewSampler: func(params []float64) (Sampler, error) {
		if idx, err := validateExponentialBeta(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		beta := params[0]
		return continuousSampler{fill: func(engine *sim.SimulatorEngine, buffer []float64) error {
			return engine.FillExponential(buffer, 1.0/beta)
		}}, nil
	},
	FormatParams: func(params []float64) string {
		if len(params) >= 1 {
			return fmt.Sprintf("β=%.4f", params[0])
		}
		return ""
	},
}

var uniformSpec = Spec{
	ID:          IDUniform,
	DisplayName: "Uniforme continua",
	Discrete:    false,
	ParamLabels: []string{"Límite inferior (a)", "Límite superior (b)"},
	Validate:    validateUniform,
	Construct: func(params []float64) (distributions.Distribution, error) {
		if idx, err := validateUniform(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		return distributions.NewUniform(params[0], params[1])
	},
	NewSampler: func(params []float64) (Sampler, error) {
		if idx, err := validateUniform(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		a, b := params[0], params[1]
		return continuousSampler{fill: func(engine *sim.SimulatorEngine, buffer []float64) error {
			return engine.FillUniformContinuous(buffer, a, b)
		}}, nil
	},
	FormatParams: func(params []float64) string {
		if len(params) >= 2 {
			return fmt.Sprintf("a=%.4f, b=%.4f", params[0], params[1])
		}
		return ""
	},
}
