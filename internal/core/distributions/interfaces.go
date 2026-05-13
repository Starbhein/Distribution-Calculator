package distributions

type Distribution interface {
	Avg() float64
	Variance() float64
	StdDev() float64
}
type DiscreteDistribution interface {
	Distribution
	PMF(k int) float64
	PDF(k int) float64
}
type ContinuousDistribution interface {
	Distribution
	PDF(x float64) float64
	CDF(x float64) float64
}
