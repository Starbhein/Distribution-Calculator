package distributions

import (
	"errors"
	"math"
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
		return 0.0, errors.New("the k constant couldn't be negative, it has to be greater or equals than 0. ")
	}
	num := float64(k)*math.Log(p.Lambda) - p.Lambda
	den, _ := math.Lgamma(float64(k + 1))
	return math.Exp(num - den), nil

	return 0, nil
}

func (p Poisson) CDF(k int) (float64, error) {
	return 0, nil
}
