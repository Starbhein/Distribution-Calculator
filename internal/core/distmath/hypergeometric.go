// Package distmath is the unified mathematical core for distribution
// processing (design §2.1). PR2 delivers only the hypergeometric PMF
// surface (design §3.1); CDF kernels land in PR3a.
package distmath

import "math"

// hypergeometricSupport returns the support bounds [startK, maxK] of the
// hypergeometric distribution with M successes in a population of N and a
// sample of size nSample: max(0, nSample-(N-M)) to min(nSample, M).
func hypergeometricSupport(m, nSample, n int) (startK, maxK int) {
	startK = 0
	if lower := nSample - (n - m); lower > 0 {
		startK = lower
	}
	maxK = nSample
	if m < nSample {
		maxK = m
	}
	return startK, maxK
}

// hypergeometricMode returns the mode clamped to the support (design §3.1):
// floor((nSample+1)(M+1)/(N+2)) — integer division is the floor for
// non-negative operands.
func hypergeometricMode(m, nSample, n int) int {
	startK, maxK := hypergeometricSupport(m, nSample, n)
	mode := ((nSample + 1) * (m + 1)) / (n + 2)
	if mode < startK {
		mode = startK
	}
	if mode > maxK {
		mode = maxK
	}
	return mode
}

// logHypergeometricPMF is the exact 9-term Lgamma log-PMF — the single seed
// evaluation per computation path (design §3.1). It is identical to the
// expression proven in the old Hypergeometric.PMF and the sim table builder.
func logHypergeometricPMF(m, nSample, n, k int) float64 {
	mF, _ := math.Lgamma(float64(m + 1))
	kF, _ := math.Lgamma(float64(k + 1))
	mMinusKF, _ := math.Lgamma(float64(m - k + 1))
	nMinusMF, _ := math.Lgamma(float64(n - m + 1))
	nSampleMinusKF, _ := math.Lgamma(float64(nSample - k + 1))
	nMinusMMinusSamplePlusKF, _ := math.Lgamma(float64(n - m - nSample + k + 1))
	nSampleF, _ := math.Lgamma(float64(nSample + 1))
	nMinusSampleF, _ := math.Lgamma(float64(n - nSample + 1))
	nF, _ := math.Lgamma(float64(n + 1))
	return mF - kF - mMinusKF + nMinusMF - nSampleMinusKF - nMinusMMinusSamplePlusKF + nSampleF + nMinusSampleF - nF
}

// hypergeometricRatio returns P(k)/P(k-1) for the hypergeometric PMF
// (design §3.1): (nSample−k+1)(M−k+1) / (k(N−M−nSample+k)).
// Callers guarantee startK < k <= maxK, so numerator and denominator are
// both strictly positive.
func hypergeometricRatio(m, nSample, n, k int) float64 {
	numerator := float64((nSample - k + 1) * (m - k + 1))
	denominator := float64(k * (n - m - nSample + k))
	return numerator / denominator
}

// HypergeometricPMF returns the hypergeometric point probability P(X = k)
// via a mode-anchored recurrence: one log-space seed at the mode plus a
// linear-space ratio walk to k (design §3.1). Points outside the support
// return exactly 0.
func HypergeometricPMF(m, nSample, n, k int) float64 {
	startK, maxK := hypergeometricSupport(m, nSample, n)
	if k < startK || k > maxK {
		return 0
	}
	mode := hypergeometricMode(m, nSample, n)
	p := math.Exp(logHypergeometricPMF(m, nSample, n, mode))
	switch {
	case k > mode:
		for i := mode + 1; i <= k; i++ {
			p *= hypergeometricRatio(m, nSample, n, i)
		}
	case k < mode:
		for i := mode; i > k; i-- {
			p /= hypergeometricRatio(m, nSample, n, i)
		}
	}
	return p
}

// HypergeometricPMFRow returns the full PMF row over the support
// [startK, maxK] using one seed plus an O(range) ratio walk, materialized
// and normalized so the row sums to 1 within ~1 ulp (design §2.1/§3.1).
// This is the support-scan path all multi-point callers (charts, tables)
// must use: it replaces range × 9-Lgamma closed-form evaluations.
func HypergeometricPMFRow(m, nSample, n int) (row []float64, startK, maxK int) {
	startK, maxK = hypergeometricSupport(m, nSample, n)
	row = make([]float64, maxK-startK+1)
	mode := hypergeometricMode(m, nSample, n)
	seed := math.Exp(logHypergeometricPMF(m, nSample, n, mode))
	row[mode-startK] = seed
	// Walk down from the mode: P(k-1) = P(k) / ratio(k).
	for k := mode; k > startK; k-- {
		row[k-1-startK] = row[k-startK] / hypergeometricRatio(m, nSample, n, k)
	}
	// Walk up from the mode: P(k) = P(k-1) * ratio(k).
	for k := mode + 1; k <= maxK; k++ {
		row[k-startK] = row[k-1-startK] * hypergeometricRatio(m, nSample, n, k)
	}
	// Normalize: the seed+walk drift (~1 ulp per step, design §2.2) is
	// absorbed here so the materialized row is an exact distribution.
	var sum float64
	for _, v := range row {
		sum += v
	}
	for i := range row {
		row[i] /= sum
	}
	return row, startK, maxK
}
