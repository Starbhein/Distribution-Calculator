package distributions

import (
	"errors"
	"math"
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

func (h Hypergeometric) PMF(k int) (float64, error) {
	if k < 0 || k > h.N {
		return 0.0, errors.New("the k constant couldn't be negative, it has to be greater or equals than 0 and lower than N ")
	}
	mFactorial, _ := math.Lgamma(float64(h.M + 1))
	kFactorial, _ := math.Lgamma(float64(k + 1))
	mminkFactorial, _ := math.Lgamma(float64(h.M - k + 1))
	nminmFactorial, _ := math.Lgamma(float64(h.N - h.M + 1))
	nminkFactorial, _ := math.Lgamma(float64(h.NSample - k + 1))
	nminmnk, _ := math.Lgamma(float64(h.N - h.M - h.NSample + k + 1))
	nsampleFactorial, _ := math.Lgamma(float64(h.NSample + 1))
	nminsampleFactorial, _ := math.Lgamma(float64(h.N - h.NSample + 1))
	nFactorial, _ := math.Lgamma(float64(h.N + 1))
	res := mFactorial - kFactorial - mminkFactorial + nminmFactorial - nminkFactorial - nminmnk + nsampleFactorial + nminsampleFactorial - nFactorial
	return math.Exp(res), nil
}

func (h Hypergeometric) CDF(k int) (float64, error) {
	if k > h.NSample {
		return 0, errors.New("k must be lower or equals than the sample's size")
	}
	if k < 0 {
		return 0, errors.New("k must be greater than 0")
	}
	var res float64
	cumulative, err := h.PMF(0)
	if err != nil {
		return 0, err
	}
	if k == 0 {
		return cumulative, nil
	}
	res = cumulative
	for i := 1; i <= k; i++ {
		cumulative *= float64(h.M-i+1) / float64(i) * float64(h.NSample-i+1) / float64(h.N-h.M-h.NSample+i)
		res += cumulative
	}

	return res, nil
}
