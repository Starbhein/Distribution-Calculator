package registry

import (
	"errors"
	"fmt"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
	"github.com/Starbhein/DistCalc/internal/core/distributions/sim"
)

// Hypergeometric spec entry (PR4a-2 scope, design §4). This file is the
// ONLY place the UI parameter order (N, M, n) is mapped to the engine
// order (m, nsample, n): the named binding below single-sources the
// mapping so a swapped argument is a test-time failure, not a silent
// wrong answer (spec §7).

// HypergeometricParams is the single-sourced named binding between the UI
// parameter order and the engine order (design §4). Every downstream
// consumer references fields, never indices.
type HypergeometricParams struct {
	Population int // N — UI order position 0
	Successes  int // M — UI order position 1
	Sample     int // n — UI order position 2
}

// bindHypergeometric validates the raw UI-ordered params and binds them to
// named fields. It is the only place params[0..2] are read for the
// hypergeometric distribution.
func bindHypergeometric(params []float64) (HypergeometricParams, error) {
	if idx, err := validateHypergeometric(params); err != nil {
		return HypergeometricParams{}, fmt.Errorf("param %d: %w", idx, err)
	}
	return HypergeometricParams{
		Population: int(params[0]),
		Successes:  int(params[1]),
		Sample:     int(params[2]),
	}, nil
}

// validateHypergeometric mirrors ui.ValidateParams, INCLUDING the
// cross-parameter rules M<=N and n<=N (design §7 backstop): rejecting
// those combos here makes the latent HypergeometricCDFTable startK>maxK
// panic unreachable through the registry path.
func validateHypergeometric(params []float64) (int, error) {
	if len(params) < 3 {
		return -1, errors.New("faltan parámetros para Hypergeométrica")
	}
	N, M, n := params[0], params[1], params[2]
	if N <= 0 {
		return 0, errors.New("N debe ser mayor que 0")
	}
	if M <= 0 {
		return 1, errors.New("M debe ser mayor que 0")
	}
	if n <= 0 {
		return 2, errors.New("n debe ser mayor que 0")
	}
	if M > N {
		return 1, errors.New("M no puede ser mayor que N")
	}
	if n > N {
		return 2, errors.New("n no puede ser mayor que N")
	}
	return -1, nil
}

// hypergeometricSampler binds HypergeometricParams once; Prebuild gates the
// CDF table on the struct's Variance() and Fill maps the named fields to
// the engine order (m, nsample, n).
type hypergeometricSampler struct {
	params HypergeometricParams
	table  []float64
}

func (s *hypergeometricSampler) Prebuild() error {
	h, err := distributions.NewHypergeometric(s.params.Successes, s.params.Population, s.params.Sample)
	if err != nil {
		return err
	}
	// Gate on the struct's Variance() (design §2.4 — kills the 4th inline
	// copy of the formula). The negated form is polarity-identical to
	// sim.HypergeometricUsesTable: a NaN variance (degenerate N==1
	// support) falls through to the table path, as before.
	if !(h.Variance() > 9.0) {
		table, _, _, err := sim.BuildHypergeometricCDFTable(
			float64(s.params.Successes), float64(s.params.Sample), float64(s.params.Population))
		if err != nil {
			return err
		}
		s.table = table
	}
	return nil
}

func (s *hypergeometricSampler) Fill(engine *sim.SimulatorEngine, buffer []float64) error {
	// Engine order (m, nsample, n) = (Successes, Sample, Population).
	return engine.FillHypergeometric(buffer,
		float64(s.params.Successes), float64(s.params.Sample), float64(s.params.Population), s.table)
}

var hypergeometricSpec = Spec{
	ID:          IDHypergeometric,
	DisplayName: "Hypergeométrica",
	Discrete:    true,
	ParamLabels: []string{"Tamaño poblacional (N)", "Número de exitos (M)", "Tamaño de muestra (n)"},
	Validate:    validateHypergeometric,
	Construct: func(params []float64) (distributions.Distribution, error) {
		p, err := bindHypergeometric(params)
		if err != nil {
			return nil, err
		}
		// Constructor order (M, N, n) = (Successes, Population, Sample).
		return distributions.NewHypergeometric(p.Successes, p.Population, p.Sample)
	},
	NewSampler: func(params []float64) (Sampler, error) {
		p, err := bindHypergeometric(params)
		if err != nil {
			return nil, err
		}
		return &hypergeometricSampler{params: p}, nil
	},
	FormatParams: func(params []float64) string {
		if len(params) >= 3 {
			return fmt.Sprintf("N=%.0f, M=%.0f, n=%.0f", params[0], params[1], params[2])
		}
		return ""
	},
}
