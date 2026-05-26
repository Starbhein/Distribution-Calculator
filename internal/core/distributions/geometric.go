package distributions

import (
	"errors"
	"math"
)

// Geometric represents the geometric distribution (number of trials until first success).
type Geometric struct {
	P float64
}

// NewGeometric creates a new Geometric distribution.
func NewGeometric(p float64) (*Geometric, error) {
	if p <= 0 || p > 1 {
		return nil, errors.New("the success probability must be in (0, 1]")
	}
	return &Geometric{P: p}, nil
}

func (g Geometric) Avg() float64 {
	return 1.0 / g.P
}

func (g Geometric) Variance() float64 {
	return (1.0 - g.P) / (g.P * g.P)
}

func (g Geometric) StdDev() float64 {
	return math.Sqrt(g.Variance())
}

// PMF returns the probability mass function for k ≥ 1.
// Uses log-space computation when p is small to avoid underflow.
func (g Geometric) PMF(k int) (float64, error) {
	if k < 1 {
		return 0.0, errors.New("k must be greater than or equal to 1 for Geometric")
	}
	// For small p or large k, compute in log-space to avoid underflow
	if g.P < 0.01 || k > 1000 {
		logPMF := float64(k-1)*math.Log(1.0-g.P) + math.Log(g.P)
		return math.Exp(logPMF), nil
	}
	return math.Pow(1.0-g.P, float64(k-1)) * g.P, nil
}

// PDF is an alias for PMF to satisfy the DiscreteDistribution interface.
func (g Geometric) PDF(k int) (float64, error) {
	return g.PMF(k)
}

// CDF returns the cumulative distribution function for k ≥ 0.
// Uses math.Expm1 for better precision when p is small.
func (g Geometric) CDF(k int) (float64, error) {
	if k < 0 {
		return 0.0, nil
	}
	if k == 0 {
		return 0.0, nil
	}
	// Use Expm1 for better precision: 1 - (1-p)^k = -expm1(k*log(1-p))
	logTerm := float64(k) * math.Log(1.0-g.P)
	return -math.Expm1(logTerm), nil
}
