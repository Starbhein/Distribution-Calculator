package distributions

// Distribution interface Defines the probability distributions behavior
// The operations guarantee the most efficient and precise result
type Distribution interface {
	Avg() float64
	Variance() float64
	StdDev() float64
}
type DiscreteDistribution interface {
	Distribution
	PMF(k int) (float64, error)
	PDF(k int) (float64, error)
}
type ContinuousDistribution interface {
	Distribution
	PDF(x float64) (float64, error)
	CDF(x float64) (float64, error)
}
