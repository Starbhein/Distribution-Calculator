package distmath

import "math"

// EpsilonSignificantValue is THE single truncation epsilon for every
// recurrence core in this package (design §2.3): the struct layer's 1e-15
// (poisson.go) and the sim layer's 1e-18 (simulatorengine.go) converge here
// at 1e-18 — strictly more accurate; the added tail mass to struct CDFs is
// <=1e-15 relative, two orders below the 1e-12 pinned-test tolerance, at a
// cost of <=3 extra recurrence iterations.
const EpsilonSignificantValue = 1e-18

// logBinomialPMF is the exact Lgamma log-PMF — the single seed evaluation
// per pointwise path. Identical to the expression proven in Binomial.PMF
// and Binomial.CDF (design §2.1 line-for-line preservation).
func logBinomialPMF(n int, p float64, k int) float64 {
	nFactorial, _ := math.Lgamma(float64(n + 1))
	kFactorial, _ := math.Lgamma(float64(k + 1))
	nminkFactorial, _ := math.Lgamma(float64(n - k + 1))
	factorialCoef := nFactorial - kFactorial - nminkFactorial
	return factorialCoef + float64(k)*(math.Log(p)) + math.Log(float64(1-p))*float64(n-k)
}

// binomialCDFCore is THE binomial CDF recurrence core (design §2.1): one
// mode-anchored, two-direction gated ratio walk with the pre-division
// hoisting proven in Binomial.CDF and BuildBinomialCDFTable. The anchor is
// min(int(n*p), hi) — the sim table builder's proven anchor, which design
// §2.2 makes THE core (the old struct anchor n*int(p) collapsed to 0 for
// p<1 and overflowed the walk to +Inf on inputs like n=1000, p=0.9). Two
// consumption shapes share this core: when table is non-nil it must have
// length n+1 and every visited relative PMF is stored (materialized shape);
// the walk always accumulates the relative-PMF sum over [0, hi] (pointwise
// shape — allocation-free when table is nil).
func binomialCDFCore(n int, p float64, hi int, table []float64) (maxValue int, sum float64) {
	preCalcI := (1.0 - p) / p
	preCalcR := p / (1.0 - p)
	maxValue = int(float64(n) * p)
	if maxValue > hi {
		maxValue = hi
	}
	cumulative := 1.0
	sum = 1.0
	if table != nil {
		table[maxValue] = 1.0
	}
	for i := maxValue - 1; i >= 0 && cumulative >= sum*EpsilonSignificantValue; i-- {
		cumulative *= preCalcI * (float64(i+1) / float64(n-(i+1)+1))
		sum += cumulative
		if table != nil {
			table[i] = cumulative
		}
	}
	cumulative = 1.0
	for i := maxValue + 1; i <= hi && cumulative >= sum*EpsilonSignificantValue; i++ {
		cumulative *= preCalcR * (float64(n-i+1) / float64(i))
		sum += cumulative
		if table != nil {
			table[i] = cumulative
		}
	}
	return maxValue, sum
}

// BinomialCDF returns P(X <= k) for X ~ Binomial(n, p): the allocation-free
// pointwise adapter over binomialCDFCore (design §2.1). The CDF at or above
// n is forced to exactly 1.0 (design §2.2); below 0 it is exactly 0.
func BinomialCDF(n int, p float64, k int) float64 {
	if k < 0 {
		return 0
	}
	if k >= n {
		return 1.0
	}
	maxValue, sum := binomialCDFCore(n, p, k, nil)
	return math.Exp(logBinomialPMF(n, p, maxValue)) * sum
}

// BinomialCDFTable returns the normalized CDF lookup table over [0, n]: the
// materialized adapter over the same binomialCDFCore (design §2.1),
// preserving BuildBinomialCDFTable's normalize-and-cumsum line-for-line.
// The last entry is forced to exactly 1.0.
func BinomialCDFTable(n int, p float64) []float64 {
	cdfTable := make([]float64, n+1)
	_, sum := binomialCDFCore(n, p, n, cdfTable)
	cdfTable[0] /= sum
	for i := 1; i <= n; i++ {
		cdfTable[i] = (cdfTable[i] / sum) + cdfTable[i-1]
	}
	cdfTable[n] = 1.0
	return cdfTable
}
