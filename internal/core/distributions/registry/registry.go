// Package registry is the single source of per-distribution knowledge
// (design §1.3). It imports distributions + sim and is imported by ui +
// export; no import cycles. Registration is plain package-level vars — no
// init() magic.
package registry

import (
	"github.com/Starbhein/DistCalc/internal/core/distributions"
	"github.com/Starbhein/DistCalc/internal/core/distributions/sim"
)

// ID is a stable internal key for a distribution ("normal", "binomial", ...).
type ID string

const (
	IDNormal            ID = "normal"
	IDExponentialLambda ID = "exponential_lambda"
	IDExponentialBeta   ID = "exponential_beta"
	IDUniform           ID = "uniform"
)

// Spec is the single source of per-distribution knowledge (design §1.3).
type Spec struct {
	ID          ID
	DisplayName string // byte-identical to the current menu strings
	Discrete    bool
	ParamLabels []string // replaces the formmodel.go:146 switch

	// Validate is THE validation layer. Returns the offending parameter
	// index (-1 for arity errors) so the UI can highlight the right field.
	Validate func(params []float64) (paramIndex int, err error)

	// Construct validates, then builds. Returns the fixed interfaces
	// (design §1.3b).
	Construct func(params []float64) (distributions.Distribution, error)

	// NewSampler binds params once; Fill is allocation-free per sample
	// (design §1.5).
	NewSampler func(params []float64) (Sampler, error)

	// FormatParams replaces the triplicated formatLLNParams /
	// formatCLTParams / plot formatters.
	FormatParams func(params []float64) string
}

// Sampler is built once per run and kills the duplicated prebuild/fill
// switch pairs (managedata.go ↔ clt.go).
type Sampler interface {
	Prebuild() error                                               // shared CDF table, once per run
	Fill(engine *sim.SimulatorEngine, buffer []float64) error      // per worker; zero per-sample allocs
}

// specs is the registry table. Discrete entries land in PR4a-2.
var specs = []Spec{
	normalSpec,
	exponentialLambdaSpec,
	exponentialBetaSpec,
	uniformSpec,
}

// specsByName is built once from specs; ByName is the ONLY name-based
// dispatch left after the refactor (design §1.3).
var specsByName = buildNameIndex(specs)

func buildNameIndex(specs []Spec) map[string]Spec {
	index := make(map[string]Spec, len(specs))
	for _, spec := range specs {
		index[spec.DisplayName] = spec
	}
	return index
}

// ByName resolves a UI menu display name to its Spec.
func ByName(displayName string) (Spec, bool) {
	spec, ok := specsByName[displayName]
	return spec, ok
}
