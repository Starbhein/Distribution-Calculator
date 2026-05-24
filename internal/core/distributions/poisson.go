package distributions

import (
	"errors"
	"math"
)

const epsilonSignificantValue = 1e-15

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

func (p Poisson) CDF(k int) (float64, error) {
	if k < 0 {
		return 0.0, errors.New("the k constant couldn't be negative, it has to be greater or equals than 0")
	}
	maxValue := min(k, int(p.Lambda))
	cumulativeR := 1.0
	sum := 1.0
	cumulativeL := 1.0
	for i := maxValue - 1; i >= 0 && cumulativeL >= sum*epsilonSignificantValue; i-- {
		cumulativeL *= float64(i+1) / p.Lambda
		sum += cumulativeL
	}
	for i := maxValue + 1; i <= k && cumulativeR >= sum*epsilonSignificantValue; i++ {
		cumulativeR *= p.Lambda / float64(i)
		sum += cumulativeR
	}

	num := float64(maxValue)*math.Log(p.Lambda) - p.Lambda
	den, _ := math.Lgamma(float64(maxValue + 1))
	return math.Exp((num - den)) * sum, nil
}
