package distributions

import (
	"errors"
	"math"
)

type Normal struct {
	Average            float64
	StdandardDeviation float64
}

func NewNormal(Average, StandardDeviation float64) (*Normal, error) {
	if StandardDeviation <= 0 {
		return nil, errors.New("the standard deviation parameter must be a positive number")
	}

	return &Normal{Average: Average, StdandardDeviation: StandardDeviation}, nil
}

func (n Normal) Avg() float64 {
	return n.Average
}

func (n Normal) StdDev() float64 {
	return n.StdandardDeviation
}

func (n Normal) Variance() float64 {
	return n.StdandardDeviation * n.StdandardDeviation
}

func (n Normal) PDF(x float64) (float64, error) {
	z := (x - n.Average) / n.StdandardDeviation
	exponent := -.5 * z * z
	coefficient := 1 / (n.StdandardDeviation * math.Sqrt(2*math.Pi))

	return coefficient * math.Exp(exponent), nil
}

func (n Normal) CDF(x float64) (float64, error) {
	z := (x - n.Average) / (n.StdandardDeviation * math.Sqrt(2))
	return 0.5 * math.Erfc(-z), nil
}
