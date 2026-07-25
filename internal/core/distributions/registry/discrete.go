package registry

import (
	"errors"
	"fmt"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
	"github.com/Starbhein/DistCalc/internal/core/distributions/sim"
)

// Discrete spec entries (PR4a-2 scope). DisplayName and ParamLabels are
// byte-identical to the current UI strings (mainmenu.go:90,
// formmodel.go:146 switch); Validate rules mirror ui.ValidateParams;
// FormatParams output is byte-identical to formatLLNParams /
// formatCLTParams. The hypergeometric entry lives in hypergeometric.go
// with its named parameter binding (design §4).

// tableSampler is the shared shape for table-backed discrete samplers:
// Prebuild materializes the shared CDF table once per run when the gating
// predicate says so (design §2.4); Fill reuses it per worker.
type tableSampler struct {
	buildTable func() ([]float64, error)
	fill       func(engine *sim.SimulatorEngine, buffer []float64, table []float64) error
	table      []float64
}

func (s *tableSampler) Prebuild() error {
	table, err := s.buildTable()
	if err != nil {
		return err
	}
	s.table = table
	return nil
}

func (s *tableSampler) Fill(engine *sim.SimulatorEngine, buffer []float64) error {
	return s.fill(engine, buffer, s.table)
}

// Validate rules mirror ui.ValidateParams exactly (design §7 convergence
// target); the UI validator dies in PR4b when consumers switch to these.
// Note the binomial UI order is (p, n) — position 0 is p, position 1 is n.
func validateBernoulli(params []float64) (int, error) {
	if len(params) < 1 {
		return -1, errors.New("faltan parámetros para Bernoulli")
	}
	if params[0] <= 0 || params[0] > 1 {
		return 0, errors.New("p debe estar en (0, 1]")
	}
	return -1, nil
}

func validateBinomial(params []float64) (int, error) {
	if len(params) < 2 {
		return -1, errors.New("faltan parámetros para Binomial")
	}
	if params[0] <= 0 || params[0] > 1 {
		return 0, errors.New("p debe estar en (0, 1]")
	}
	if params[1] <= 0 {
		return 1, errors.New("n debe ser mayor que 0")
	}
	return -1, nil
}

func validateGeometric(params []float64) (int, error) {
	if len(params) < 1 {
		return -1, errors.New("faltan parámetros para Geométrica")
	}
	if params[0] <= 0 || params[0] > 1 {
		return 0, errors.New("p debe estar en (0, 1]")
	}
	return -1, nil
}

func validatePoisson(params []float64) (int, error) {
	if len(params) < 1 {
		return -1, errors.New("faltan parámetros para Poisson")
	}
	if params[0] <= 0 {
		return 0, errors.New("λ debe ser mayor que 0")
	}
	return -1, nil
}

var bernoulliSpec = Spec{
	ID:          IDBernoulli,
	DisplayName: "Bernoulli",
	Discrete:    true,
	ParamLabels: []string{"Probabilidad (p)"},
	Validate:    validateBernoulli,
	Construct: func(params []float64) (distributions.Distribution, error) {
		if idx, err := validateBernoulli(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		return distributions.NewBernoulli(params[0])
	},
	NewSampler: func(params []float64) (Sampler, error) {
		if idx, err := validateBernoulli(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		p := params[0]
		// Table-less sampler: no-op Prebuild (design §1.5).
		return continuousSampler{fill: func(engine *sim.SimulatorEngine, buffer []float64) error {
			return engine.FillBernoulli(buffer, p)
		}}, nil
	},
	FormatParams: func(params []float64) string {
		if len(params) >= 1 {
			return fmt.Sprintf("p=%.4f", params[0])
		}
		return ""
	},
}

var binomialSpec = Spec{
	ID:          IDBinomial,
	DisplayName: "Binomial",
	Discrete:    true,
	ParamLabels: []string{"Probabilidad (p)", "Ensayos (N)"},
	Validate:    validateBinomial,
	Construct: func(params []float64) (distributions.Distribution, error) {
		if idx, err := validateBinomial(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		// UI order (p, n) → constructor order (n, p).
		return distributions.NewBinomial(int(params[1]), params[0])
	},
	NewSampler: func(params []float64) (Sampler, error) {
		if idx, err := validateBinomial(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		p, n := params[0], int(params[1])
		return &tableSampler{
			buildTable: func() ([]float64, error) {
				if sim.BinomialUsesTable(n, p) {
					return sim.BuildBinomialCDFTable(n, p), nil
				}
				return nil, nil
			},
			fill: func(engine *sim.SimulatorEngine, buffer []float64, table []float64) error {
				return engine.FillBinomial(buffer, n, p, table)
			},
		}, nil
	},
	FormatParams: func(params []float64) string {
		if len(params) >= 2 {
			return fmt.Sprintf("p=%.4f, n=%.0f", params[0], params[1])
		}
		return ""
	},
}

var geometricSpec = Spec{
	ID:          IDGeometric,
	DisplayName: "Geométrica",
	Discrete:    true,
	ParamLabels: []string{"Probabilidad (p)"},
	Validate:    validateGeometric,
	Construct: func(params []float64) (distributions.Distribution, error) {
		if idx, err := validateGeometric(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		return distributions.NewGeometric(params[0])
	},
	NewSampler: func(params []float64) (Sampler, error) {
		if idx, err := validateGeometric(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		p := params[0]
		// Table-less sampler (O(1) inverse-CDF fill): no-op Prebuild.
		return continuousSampler{fill: func(engine *sim.SimulatorEngine, buffer []float64) error {
			return engine.FillGeometric(buffer, p)
		}}, nil
	},
	FormatParams: func(params []float64) string {
		if len(params) >= 1 {
			return fmt.Sprintf("p=%.4f", params[0])
		}
		return ""
	},
}

var poissonSpec = Spec{
	ID:          IDPoisson,
	DisplayName: "Poisson",
	Discrete:    true,
	ParamLabels: []string{"Lambda (λ)"},
	Validate:    validatePoisson,
	Construct: func(params []float64) (distributions.Distribution, error) {
		if idx, err := validatePoisson(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		return distributions.NewPoisson(params[0])
	},
	NewSampler: func(params []float64) (Sampler, error) {
		if idx, err := validatePoisson(params); err != nil {
			return nil, fmt.Errorf("param %d: %w", idx, err)
		}
		lambda := params[0]
		return &tableSampler{
			buildTable: func() ([]float64, error) {
				if sim.PoissonUsesTable(lambda) {
					return sim.BuildPoissonCDFTable(lambda), nil
				}
				return nil, nil
			},
			fill: func(engine *sim.SimulatorEngine, buffer []float64, table []float64) error {
				return engine.FillPoisson(buffer, lambda, table)
			},
		}, nil
	},
	FormatParams: func(params []float64) string {
		if len(params) >= 1 {
			return fmt.Sprintf("λ=%.4f", params[0])
		}
		return ""
	},
}
