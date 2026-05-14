package distributions

type Poisson struct {
	Lambda float64
}

func (p Poisson) Avg() float64 {
	return 0
}

func (p Poisson) Variance() float64 {
	return 0
}

func (p Poisson) StdDev() float64 {
	return 0
}
