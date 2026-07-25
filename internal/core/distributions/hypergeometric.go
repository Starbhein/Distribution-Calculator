package distributions

import (
	"errors"
	"math"

	"github.com/Starbhein/DistCalc/internal/core/distmath"
)

type Hypergeometric struct {
	M       int
	N       int
	NSample int
}

func NewHypergeometric(successQuantity, populationQuantity, sampleQuantity int) (*Hypergeometric, error) {
	if successQuantity <= 0 {
		return nil, errors.New("the population success quantity must be greater than 0 ")
	}
	if populationQuantity <= 0 {
		return nil, errors.New("the population quantity must be greater than 0")
	}
	if sampleQuantity <= 0 {
		return nil, errors.New("the sample quantity must be greater than 0")
	}
	// Cross-parameter backstops (design §7, spec §3): defense for direct
	// library consumers. The registry Validate layer owns the app-level
	// rule; these keep the two layers from ever disagreeing.
	if successQuantity > populationQuantity {
		return nil, errors.New("the population success quantity must not exceed the population quantity")
	}
	if sampleQuantity > populationQuantity {
		return nil, errors.New("the sample quantity must not exceed the population quantity")
	}
	return &Hypergeometric{M: successQuantity, N: populationQuantity, NSample: sampleQuantity}, nil
}

func (h Hypergeometric) Avg() float64 {
	return float64(h.NSample) * float64(h.M) / float64(h.N)
}

func (h Hypergeometric) Variance() float64 {
	return float64(h.NSample) * (float64(h.M) / float64(h.N)) * float64(h.N-h.M) / float64(h.N) * float64(h.N-h.NSample) / float64(h.N-1)
}

func (h Hypergeometric) StdDev() float64 {
	return math.Sqrt(h.Variance())
}

// PMF delegates to the distmath mode-anchored recurrence (design §3.1):
// one log-space seed plus a short ratio walk instead of 9 Lgamma calls
// per point. The k guard is preserved exactly (no behavior change).
func (h Hypergeometric) PMF(k int) (float64, error) {
	if k < 0 || k > h.N {
		return 0.0, errors.New("the k constant couldn't be negative, it has to be greater or equals than 0 and lower than N ")
	}
	return distmath.HypergeometricPMF(h.M, h.NSample, h.N, k), nil
}

// CDF delegates to the distmath pointwise kernel (design §2.1): the same
// mode-anchored recurrence core, allocation-free per call. The k guards are
// preserved exactly (no behavior change).
func (h Hypergeometric) CDF(k int) (float64, error) {
	if k > h.NSample {
		return 0, errors.New("k must be lower or equals than the sample's size")
	}

	if k < 0 {
		return 0, errors.New("k must be greater than 0")
	}
	return distmath.HypergeometricCDF(h.M, h.NSample, h.N, k), nil
}
