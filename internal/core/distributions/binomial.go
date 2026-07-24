// Package distributions it's the main operation package, It describes the different distribution approachs that the calculator has
package distributions

import (
	"errors"
	"math"

	"github.com/Starbhein/DistCalc/internal/core/distmath"
)

// Binomial it's the defined package to simulate the distribution probabilistic model
// If the sample size/population size >=0.05, the experiment isn't binomial
type Binomial struct {
	N  int
	EP float64
}

func NewBinomial(n int, ep float64) (*Binomial, error) {
	if ep > 1 || ep < 0 {
		return nil, errors.New("the success probability must be between 1 and 0")
	}
	if n <= 0 {
		return nil, errors.New("'n' Must be greater than 0")
	}
	res := &Binomial{N: n, EP: ep}
	return res, nil
}

func (b Binomial) Variance() float64 {
	return float64(b.N) * b.EP * (1 - b.EP)
}

func (b Binomial) Avg() float64 {
	return float64(b.N) * b.EP
}

func (b Binomial) StdDev() float64 {
	return math.Sqrt(b.Variance())
}

func (b Binomial) PMF(k int) (float64, error) {
	if k == 0 {
		return math.Exp(float64(b.N) * math.Log((1 - b.EP))), nil
	}
	if k < 0 || k > b.N {
		return 0.0, errors.New("the k constant couldn't be negative, it has to be greater or equals than 0 and lower than N ")
	}
	nFactorial, _ := math.Lgamma(float64(b.N + 1))
	kFactorial, _ := math.Lgamma(float64(k + 1))           // should be substract on the final operation
	nminkFactorial, _ := math.Lgamma(float64(b.N - k + 1)) // should be substract on the final operation
	factorialCoef := nFactorial - kFactorial - nminkFactorial
	res := factorialCoef + float64(k)*(math.Log(b.EP)) + math.Log(float64(1-b.EP))*float64(b.N-k)
	return math.Exp(res), nil
}

// CDF delegates to the distmath pointwise kernel (design §2.1): the same
// recurrence core the sim table builder uses, allocation-free per call.
// The k guards are preserved exactly (no behavior change).
func (b Binomial) CDF(k int) (float64, error) {
	if k == 0 {
		return math.Exp(float64(b.N) * math.Log((1 - b.EP))), nil
	}

	if k < 0 || k > b.N {
		return 0.0, errors.New("the k constant couldn't be negative, it has to be greater or equals than 0 and lower than N")
	}
	return distmath.BinomialCDF(b.N, b.EP, k), nil
}
