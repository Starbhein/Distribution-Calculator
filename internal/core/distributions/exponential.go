package distributions

import (
	"errors"
	"math"
)

// Exponential struct represents the exponential distribution
type ExponentialLambda struct {
	Lambda float64
}

func NewExponentialLambda(lambda float64) (*ExponentialLambda, error) {
	if lambda < 0 {
		return nil, errors.New("lambda must be greater than 0")
	}
	return &ExponentialLambda{Lambda: lambda}, nil
}

func (el ExponentialLambda) Avg() float64 {
	return float64(1) / el.Lambda
}

func (el ExponentialLambda) Variance() float64 {
	return math.Pow(float64(1)/el.Lambda, 2)
}

func (el ExponentialLambda) StdDev() float64 {
	return float64(1) / el.Lambda
}

func (el ExponentialLambda) PDF(x float64) (float64, error) {
	if x < 0 {
		return 0.0, errors.New("x must be a positive number ")
	}
	return el.Lambda * math.Exp(-el.Lambda*x), nil
}

func (el ExponentialLambda) CDF(x float64) (float64, error) {
	if x < 0 {
		return 0.0, errors.New("x must be a positive number ")
	}
	return 1 - math.Exp(-el.Lambda*x), nil
}

type ExponentialBetha struct {
	Beta float64
}

func NewExponentialBetha(betha float64) (*ExponentialBetha, error) {
	if betha < 0 {
		return nil, errors.New("betha must be greater than 0")
	}
	return &ExponentialBetha{Beta: betha}, nil
}

func (eb ExponentialBetha) Avg() float64 {
	return eb.Beta
}

func (eb ExponentialBetha) Variance() float64 {
	return math.Pow(eb.Beta, 2)
}

func (eb ExponentialBetha) StdDev() float64 {
	return eb.Beta
}

func (eb ExponentialBetha) PDF(x float64) (float64, error) {
	if x < 0 {
		return 0.0, errors.New("x must be a positive number ")
	}
	if x == 0 {
		return 1, nil
	}
	return (float64(1) / eb.Beta) * math.Exp(-(float64(1)/eb.Beta)*x), nil
}

func (eb ExponentialBetha) CDF(x float64) (float64, error) {
	if x < 0 {
		return 0.0, errors.New("x must be a positive number ")
	}
	return float64(1) - math.Exp(-(float64(1)/eb.Beta)*x), nil
}
