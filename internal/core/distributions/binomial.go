package distributions

// package distributions it's the main operation package, It describes the different distribution approachs that the calculator has

type Binomial struct {
	N  int
	EP float64
}

func (b Binomial) Variance() (float64, error) {
	return float64(b.N) * b.EP * (1 - b.EP), nil
}

func (b Binomial) Avg() (float64, error) {
	return float64(b.N) * b.EP, nil
}

func (b Binomial) Prob(k int) (float64, error) {
	return 0, nil
}
