package sim

import (
	"errors"
	"math"
	"math/rand/v2"
	"sort"

	"github.com/Starbhein/DistCalc/internal/core/distmath"
)

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
// It delegates to the unified distmath kernel (design §2.1); the sim-local
// recurrence and epsilonSignificantEpsilon are gone — distmath.EpsilonSignificantValue
// is THE single truncation epsilon (design §2.3).
func BuildBinomialCDFTable(n int, success float64) []float64 {
	return distmath.BinomialCDFTable(n, success)
}

func (engine *SimulatorEngine) FillBinomial(buffer []float64, n int, success float64, cdfTable []float64) error {
	if !BinomialUsesTable(n, success) {
		variance := float64(n) * success * (1.0 - success)
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
// It delegates to the unified distmath kernel (design §2.1).
func BuildPoissonCDFTable(lambda float64) []float64 {
	return distmath.PoissonCDFTable(lambda)
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
// It delegates to the unified distmath kernel (design §2.1).
func BuildHypergeometricCDFTable(m, nsample, n float64) ([]float64, float64, float64, error) {
	table, startK, maxK, err := distmath.HypergeometricCDFTable(int(m), int(nsample), int(n))
	if err != nil {
		return nil, 0, 0, err
	}
	return table, float64(startK), float64(maxK), nil
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

	// Gate on the single-sourced predicate (design §2.4); the variance itself
	// is still needed for the normal path's stdDev.
	if !HypergeometricUsesTable(m, nsample, n) {
		variance := hypergeometricVariance(M, K, N)
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

// FillGeometric uses the exact inverse-CDF of the trials-until-success
// geometric distribution for ALL p (design §3.2): k = ceil(log(u)/log1p(-p)).
// math.Log1p preserves full precision at small p, so the old O(1/p)
// iterative small-p branch is deleted — one PRNG draw per sample.
func (engine *SimulatorEngine) FillGeometric(buffer []float64, p float64) error {
	log1mp := math.Log1p(-p)
	for i := range buffer {
		u := engine.prng.Float64()
		if u == 0 {
			// P ≈ 2^-53 guard: log(0) = -Inf would yield +Inf.
			u = math.SmallestNonzeroFloat64
		}
		k := math.Ceil(math.Log(u) / log1mp)
		if k < 1 {
			k = 1
		}
		buffer[i] = k
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
