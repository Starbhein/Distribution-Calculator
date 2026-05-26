package sim

import (
	"errors"
	"math"
	"math/rand/v2"
	"sort"
)

const epsilonSignificantEpsilon = 1e-18

type SimulatorEngine struct {
	prng *rand.Rand
}

// NewSimulatorEngine uses PRNG via PCG (Permuted Congruential Generator) to simulate randomized data
func NewSimulatorEngine(seed1, seed2 uint64) *SimulatorEngine {
	source := rand.NewPCG(seed1, seed2)
	return &SimulatorEngine{prng: rand.New(source)}
}

// BuildBinomialCDFTable constructs the CDF lookup table for binomial simulation.
// Callers can pre-build and share this table across workers.
func BuildBinomialCDFTable(n int, success float64) []float64 {
	preCalcI := (1.0 - success) / success
	preCalcR := success / (1.0 - success)
	cdfTable := make([]float64, n+1)
	maxValue := int(float64(n) * success)
	cumulative := 1.0
	sum := 1.0
	cdfTable[maxValue] = 1.0
	for i := maxValue - 1; i >= 0 && cumulative >= sum*epsilonSignificantEpsilon; i-- {
		cumulative *= preCalcI * (float64(i+1) / float64(n-(i+1)+1))
		sum += cumulative
		cdfTable[i] = cumulative
	}
	cumulative = 1.0
	for i := maxValue + 1; i <= n && cumulative >= sum*epsilonSignificantEpsilon; i++ {
		cumulative *= preCalcR * (float64(n-i+1) / float64(i))
		sum += cumulative
		cdfTable[i] = cumulative
	}
	cdfTable[0] /= sum
	for i := 1; i <= n; i++ {
		cdfTable[i] = (cdfTable[i] / sum) + cdfTable[i-1]
	}
	cdfTable[n] = 1.0
	return cdfTable
}

func (engine *SimulatorEngine) FillBinomial(buffer []float64, n int, success float64, cdfTable []float64) error {
	variance := float64(n) * success * (1.0 - success)
	if variance > 9.0 {
		stdDev := math.Sqrt(variance)
		avg := float64(n) * success
		for i := range buffer {
			z := engine.prng.NormFloat64()
			val := math.Round(avg + z*stdDev)
			if val < 0 {
				val = 0
			}
			if val > float64(n) {
				val = float64(n)
			}
			buffer[i] = val
		}
		return nil
	}
	if cdfTable == nil {
		cdfTable = BuildBinomialCDFTable(n, success)
	}
	for i := range buffer {
		u := engine.prng.Float64()
		res := sort.Search(len(cdfTable), func(j int) bool { return cdfTable[j] >= u })
		buffer[i] = float64(res)
	}
	return nil
}

// BuildPoissonCDFTable constructs the CDF lookup table for Poisson simulation.
// Size is approximately lambda + 4*sqrt(lambda), which captures >99.99% of probability mass.
func BuildPoissonCDFTable(lambda float64) []float64 {
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

func (engine *SimulatorEngine) FillPoisson(buffer []float64, lambda float64, cdfTable []float64) error {
	if lambda <= 0 {
		return errors.New("lambda must be greater than 0")
	}
	if lambda > 100 {
		avg := lambda
		stdDev := math.Sqrt(lambda)
		for i := range buffer {
			z := engine.prng.NormFloat64()
			val := math.Round(avg + z*stdDev)
			if val < 0 {
				val = 0
			}
			buffer[i] = val
		}
		return nil
	}
	// For medium lambda (10-100), use CDF table + binary search
	if lambda > 10 {
		if cdfTable == nil {
			cdfTable = BuildPoissonCDFTable(lambda)
		}
		for i := range buffer {
			u := engine.prng.Float64()
			res := sort.Search(len(cdfTable), func(j int) bool { return cdfTable[j] >= u })
			buffer[i] = float64(res)
		}
		return nil
	}
	// For small lambda (<= 10), iterative method is faster than table + binary search
	for i := range buffer {
		u := engine.prng.Float64()
		p := math.Exp(-lambda)
		f := p
		k := 0
		for u > f {
			k++
			p *= lambda / float64(k)
			f += p
		}
		buffer[i] = float64(k)
	}
	return nil
}

// BuildHypergeometricCDFTable constructs the CDF lookup table for hypergeometric simulation.
// Parameters are float64 for API compatibility but converted to int internally.
func BuildHypergeometricCDFTable(m, nsample, n float64) ([]float64, float64, float64, error) {
	if nsample > n {
		return nil, 0, 0, errors.New("the sample size must be lower than the N population size")
	}
	// Convert to int for internal precision
	M := int(m)
	N := int(n)
	K := int(nsample)
	startK := 0
	if K > (N - M) {
		startK = K - (N - M)
	}
	maxK := K
	if M < K {
		maxK = M
	}
	tableSize := maxK - startK + 1
	cdfTable := make([]float64, tableSize)

	// Compute initial PDF at startK using log-gamma
	gammaM, _ := math.Lgamma(m + 1)
	gammaK, _ := math.Lgamma(float64(startK) + 1)
	gammaMminK, _ := math.Lgamma(m - float64(startK) + 1)
	gammaNminM, _ := math.Lgamma(n - m + 1)
	gammaSampleMinK, _ := math.Lgamma(nsample - float64(startK) + 1)
	gammaNminMminSamplePlusK, _ := math.Lgamma(n - m - nsample + float64(startK) + 1)
	gammaNf, _ := math.Lgamma(n + 1)
	gammaSamplef, _ := math.Lgamma(nsample + 1)
	gammaNminSample, _ := math.Lgamma(n - nsample + 1)

	res := (gammaM - gammaK - gammaMminK) +
		(gammaNminM - gammaSampleMinK - gammaNminMminSamplePlusK) -
		(gammaNf - gammaSamplef - gammaNminSample)
	initialPdf := math.Exp(res)

	// Build PMF values from startK to maxK
	pdf := initialPdf
	sum := 0.0
	for idx, k := 0, startK; k <= maxK; idx, k = idx+1, k+1 {
		if idx > 0 {
			pdf *= float64((K-k+1)*(M-k+1)) / float64(k*(N-M-K+k))
		}
		sum += pdf
		cdfTable[idx] = sum
	}
	// Normalize
	for i := range cdfTable {
		cdfTable[i] /= sum
	}
	cdfTable[tableSize-1] = 1.0
	return cdfTable, float64(startK), float64(maxK), nil
}

func (engine *SimulatorEngine) FillHypergeometric(buffer []float64, m, nsample, n float64, cdfTable []float64) error {
	if nsample > n {
		return errors.New("the sample size must be lower than the N population size")
	}
	// Convert to int for all internal calculations
	M := int(m)
	N := int(n)
	K := int(nsample)
	startK := 0
	if K > (N - M) {
		startK = K - (N - M)
	}
	maxK := K
	if M < K {
		maxK = M
	}

	// Compute variance for normal approximation check
	// Variance = K * (M/N) * ((N-M)/N) * ((N-K)/(N-1))
	variance := float64(K) * (float64(M) / float64(N)) * (float64(N-M) / float64(N)) * (float64(N-K) / float64(N-1))
	if variance > 9.0 {
		mean := float64(K) * float64(M) / float64(N)
		stdDev := math.Sqrt(variance)
		for i := range buffer {
			z := engine.prng.NormFloat64()
			val := math.Round(mean + z*stdDev)
			if val < float64(startK) {
				val = float64(startK)
			}
			if val > float64(maxK) {
				val = float64(maxK)
			}
			buffer[i] = val
		}
		return nil
	}

	if cdfTable == nil {
		var err error
		cdfTable, _, _, err = BuildHypergeometricCDFTable(m, nsample, n)
		if err != nil {
			return err
		}
	}
	for i := range buffer {
		u := engine.prng.Float64()
		res := sort.Search(len(cdfTable), func(j int) bool { return cdfTable[j] >= u })
		buffer[i] = float64(startK + res)
	}
	return nil
}

func (engine *SimulatorEngine) FillNormal(buffer []float64, avg, stdDev float64) error {
	for i := range buffer {
		buffer[i] = avg + engine.prng.NormFloat64()*stdDev
	}
	return nil
}

func (engine *SimulatorEngine) FillBernoulli(buffer []float64, p float64) error {
	for i := range buffer {
		if engine.prng.Float64() < p {
			buffer[i] = 1
		} else {
			buffer[i] = 0
		}
	}
	return nil
}

// FillGeometric uses a hybrid method for precision:
// - Inverse transform (log method) for p >= 0.01: O(1) per sample
// - Iterative method for p < 0.01: avoids catastrophic precision loss in log(1-p)
func (engine *SimulatorEngine) FillGeometric(buffer []float64, p float64) error {
	if p >= 0.01 {
		log1p := math.Log(1.0 - p)
		for i := range buffer {
			u := engine.prng.Float64()
			k := math.Ceil(math.Log(u) / log1p)
			if k < 1 {
				k = 1
			}
			buffer[i] = k
		}
		return nil
	}
	// Iterative method for very small p to avoid log precision loss
	for i := range buffer {
		trials := 1.0
		for engine.prng.Float64() >= p {
			trials++
		}
		buffer[i] = trials
	}
	return nil
}

func (engine *SimulatorEngine) FillUniformContinuous(buffer []float64, a, b float64) error {
	width := b - a
	for i := range buffer {
		buffer[i] = a + width*engine.prng.Float64()
	}
	return nil
}

func (engine *SimulatorEngine) FillExponential(buffer []float64, lambda float64) error {
	if lambda <= 0 {
		return errors.New("lambda must be greater than 0")
	}

	for i := range buffer {
		u := engine.prng.Float64()
		buffer[i] = -math.Log(u) / lambda
	}
	return nil
}
