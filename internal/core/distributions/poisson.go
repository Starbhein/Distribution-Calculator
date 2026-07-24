package distributions

import (
	"errors"
	"math"

	"github.com/Starbhein/DistCalc/internal/core/distmath"
)

type Poisson struct {
	Lambda float64
}

func NewPoisson(lambda float64) (*Poisson, error) {
	if lambda <= 0 {
		return nil, errors.New("lambda value couldn't be negative or equals to 0")
	}
	return &Poisson{Lambda: lambda}, nil
}

func (p Poisson) Avg() float64 {
	return p.Lambda
}

func (p Poisson) Variance() float64 {
	return p.Lambda
}

func (p Poisson) StdDev() float64 {
	return math.Sqrt(p.Lambda)
}

func (p Poisson) PMF(k int) (float64, error) {
	if k < 0 {
		return 0.0, errors.New("the k constant couldn't be negative, it has to be greater or equals than 0")
	}
	if k == 0 {
		return math.Exp(-p.Lambda), nil
	}
	num := float64(k)*math.Log(p.Lambda) - p.Lambda
	den, _ := math.Lgamma(float64(k + 1))
	return math.Exp(num - den), nil
}

// CDF delegates to the distmath pointwise kernel (design §2.1): THE
// recurrence core keeps this struct's algorithm verbatim, with the
// truncation epsilon single-sourced at distmath.EpsilonSignificantValue
// (design §2.3 — the old 1e-15 constant here is deleted, hidden coupling
// removed). The k guard is preserved exactly (no behavior change).
func (p Poisson) CDF(k int) (float64, error) {
	if k < 0 {
		return 0.0, errors.New("the k constant couldn't be negative, it has to be greater or equals than 0")
	}
	return distmath.PoissonCDF(p.Lambda, k), nil
}
