package distributions

// Distribution interface Defines the probability distributions behavior
// The operations guarantee the most efficient and precise result
type Distribution interface {
	Avg() float64
	Variance() float64
	StdDev() float64
}

// DiscreteDistribution is the fixed discrete interface (design §1.3b).
// CDF was added — the most-used operation was in no interface — and the
// misnamed PDF(k) requirement was removed (zero callers; the struct-level
// PDF aliases on Bernoulli and Geometric stay, they are just not
// interface-required anymore).
type DiscreteDistribution interface {
	Distribution
	PMF(k int) (float64, error)
	CDF(k int) (float64, error)
}

type ContinuousDistribution interface {
	Distribution
	PDF(x float64) (float64, error)
	CDF(x float64) (float64, error)
}

// Compile-time conformance assertions (design §1.3b). All structs conform
// with zero method additions. The Uniform assertion lives in uniform.go.
var (
	_ DiscreteDistribution = (*Bernoulli)(nil)
	_ DiscreteDistribution = (*Binomial)(nil)
	_ DiscreteDistribution = (*Geometric)(nil)
	_ DiscreteDistribution = (*Poisson)(nil)
	_ DiscreteDistribution = (*Hypergeometric)(nil)

	_ ContinuousDistribution = (*Normal)(nil)
	_ ContinuousDistribution = (*ExponentialLambda)(nil)
	_ ContinuousDistribution = (*ExponentialBeta)(nil)
)
