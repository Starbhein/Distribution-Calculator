package distributions

import "math"

// package distributions it's the main operation package, It describes the different distribution approachs that the calculator has

type Binomial struct {
	N  int
	EP float64
}

func (b Binomial) Variance() float64 {
	return float64(b.N) * b.EP * (1 - b.EP)
}

func (b Binomial) Avg() float64 {
	return float64(b.N) * b.EP
}

func (b Binomial) StdDev() float64 {
	return math.Sqrt(b.Avg())
}
