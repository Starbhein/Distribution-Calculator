package sim

// Gating predicates deciding CDF-table vs normal-approximation sampling,
// single-sourced here per design §2.4. They are consumed by the Fill methods
// below and (from PR4a-2 on) by the registry samplers' Prebuild, replacing
// the triplicated copies in ui/managedata.go and ui/clt.go.

// BinomialUsesTable reports whether Binomial(n, p) sampling uses the CDF
// table (true) or the normal approximation (false): the table is used iff
// the variance n*p*(1-p) <= 9. The negated form keeps the exact NaN
// polarity of the replaced inline gate (`if variance > 9.0 { normal }`):
// a NaN variance falls through to the table path, as before.
func BinomialUsesTable(n int, p float64) bool {
	return !(float64(n)*p*(1.0-p) > 9.0)
}

// PoissonUsesTable reports whether Poisson(lambda) sampling uses the CDF
// table (true) or another path (false): the table is used iff
// 10 < lambda <= 100 (lambda <= 10 uses the iterative method, lambda > 100
// the normal approximation).
func PoissonUsesTable(lambda float64) bool {
	return lambda > 10 && lambda <= 100
}

// HypergeometricUsesTable reports whether hypergeometric sampling uses the
// CDF table (true) or the normal approximation (false): the table is used
// iff the variance <= 9. Parameters are float64 for API compatibility and
// converted to int exactly as FillHypergeometric does, so the gate is
// behavior-identical to the inline check it replaces. The negated form
// keeps the exact NaN polarity of that inline gate (`if variance > 9.0
// { normal }`): degenerate supports with N==1 divide by (N-1)==0, and the
// resulting NaN variance must fall through to the table path, as before.
func HypergeometricUsesTable(m, nsample, n float64) bool {
	return !(hypergeometricVariance(int(m), int(nsample), int(n)) > 9.0)
}

// hypergeometricVariance computes K*(M/N)*((N-M)/N)*((N-K)/(N-1)) — the
// single home of the hypergeometric variance formula inside sim (design
// §2.4), shared by HypergeometricUsesTable and FillHypergeometric's normal
// path.
func hypergeometricVariance(m, nsample, n int) float64 {
	return float64(nsample) * (float64(m) / float64(n)) * (float64(n-m) / float64(n)) * (float64(n-nsample) / float64(n-1))
}
