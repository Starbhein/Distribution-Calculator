package distributions

import (
	"errors"
	"math"
)

// Bernoulli represents the Bernoulli distribution with success probability p.
type Bernoulli struct {
	P float64
}

// NewBernoulli creates a new Bernoulli distribution.
func NewBernoulli(p float64) (*Bernoulli, error) {
	if p <= 0 || p > 1 {
		return nil, errors.New("the success probability must be in (0, 1]")
	}
	return &Bernoulli{P: p}, nil
}

func (b Bernoulli) Avg() float64 {
	return b.P
}

func (b Bernoulli) Variance() float64 {
	return b.P * (1.0 - b.P)
}

func (b Bernoulli) StdDev() float64 {
	return math.Sqrt(b.Variance())
}

// PMF returns the probability mass function for k ∈ {0, 1}.
func (b Bernoulli) PMF(k int) (float64, error) {
	if k < 0 || k > 1 {
		return 0.0, errors.New("k must be 0 or 1 for Bernoulli")
	}
	if k == 0 {
		return 1.0 - b.P, nil
	}
	return b.P, nil
}

// PDF is an alias for PMF to satisfy the DiscreteDistribution interface.
func (b Bernoulli) PDF(k int) (float64, error) {
	return b.PMF(k)
}

// CDF returns the cumulative distribution function.
func (b Bernoulli) CDF(k int) (float64, error) {
	if k < 0 {
		return 0.0, nil
	}
	if k < 1 {
		return 1.0 - b.P, nil
	}
	return 1.0, nil
}
