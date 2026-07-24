package distmath

import "math"

// logPoissonPMF is the exact log-PMF seed — identical to the expression
// proven in Poisson.PMF and Poisson.CDF (design §2.1 line-for-line
// preservation). Combined with the P(k)/P(k-1) = lambda/k recurrence ratio
// and EpsilonSignificantValue it forms the poisson core both adapters
// consume.
func logPoissonPMF(lambda float64, k int) float64 {
	num := float64(k)*math.Log(lambda) - lambda
	den, _ := math.Lgamma(float64(k + 1))
	return num - den
}

// PoissonCDF returns P(X <= k) for X ~ Poisson(lambda): the allocation-free
// pointwise adapter (design §2.1). THE core is the current struct algorithm
// kept verbatim per design §2.2 (anchor at min(k, int(lambda)) with a
// two-direction gated sum); the only change is the truncation epsilon
// (1e-15 -> EpsilonSignificantValue = 1e-18, drift <=1e-15 relative, §2.3).
func PoissonCDF(lambda float64, k int) float64 {
	if k < 0 {
		return 0
	}
	maxValue := min(k, int(lambda))
	cumulativeR := 1.0
	sum := 1.0
	cumulativeL := 1.0
	for i := maxValue - 1; i >= 0 && cumulativeL >= sum*EpsilonSignificantValue; i-- {
		cumulativeL *= float64(i+1) / lambda
		sum += cumulativeL
	}
	for i := maxValue + 1; i <= k && cumulativeR >= sum*EpsilonSignificantValue; i++ {
		cumulativeR *= lambda / float64(i)
		sum += cumulativeR
	}
	return math.Exp(logPoissonPMF(lambda, maxValue)) * sum
}

// PoissonCDFTable returns the normalized CDF lookup table for Poisson
// simulation: the materialized adapter over the same core (design §2.1),
// preserving BuildPoissonCDFTable line-for-line. Size is approximately
// lambda + 4*sqrt(lambda), which captures >99.99% of probability mass; the
// table is normalized so the last entry is exactly 1.0.
func PoissonCDFTable(lambda float64) []float64 {
	stdDev := math.Sqrt(lambda)
	kMax := int(lambda+4.0*stdDev) + 1
	if kMax < 1 {
		kMax = 1
	}
	cdfTable := make([]float64, kMax+1)
	pmf := math.Exp(-lambda)
	cdfTable[0] = pmf
	for k := 1; k <= kMax; k++ {
		pmf *= lambda / float64(k)
		cdfTable[k] = cdfTable[k-1] + pmf
	}
	// Normalize to ensure last element is exactly 1.0
	last := cdfTable[kMax]
	for k := range cdfTable {
		cdfTable[k] /= last
	}
	cdfTable[kMax] = 1.0
	return cdfTable
}
