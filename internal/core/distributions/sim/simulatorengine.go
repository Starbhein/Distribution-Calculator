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

func (engine *SimulatorEngine) FillBinomial(buffer []float64, n int, success float64) error {
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
	for i := range buffer {
		u := engine.prng.Float64()
		res := sort.Search(len(cdfTable), func(j int) bool { return cdfTable[j] >= u })
		buffer[i] = float64(res)
	}
	return nil
}

func (engine *SimulatorEngine) FillPoisson(buffer []float64, lambda float64) error {
	if lambda < 0 || lambda > 600 {
		return errors.New("lambda must be between 0 and 600")
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
	// cdfTable := make([]float64, len(buffer))
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

func (engine *SimulatorEngine) FillHypergeometric(buffer []float64, m, nsample, n float64) error {
	if nsample > n {
		return errors.New("the sample size must be lower than the N population size")
	}
	startK := 0.0
	if nsample > (n - m) {
		startK = nsample - (n - m)
	}
	gammaM, _ := math.Lgamma(m + 1)
	gammaK, _ := math.Lgamma(startK + 1)
	gammaMminK, _ := math.Lgamma(m - startK + 1)
	gammaNminM, _ := math.Lgamma(n - m + 1)
	gammaSampleMinK, _ := math.Lgamma(nsample - startK + 1)
	gammaNminMminSamplePlusK, _ := math.Lgamma(n - m - nsample + startK + 1)
	gammaN, _ := math.Lgamma(n + 1)
	gammaSample, _ := math.Lgamma(nsample + 1)
	gammaNminSample, _ := math.Lgamma(n - nsample + 1)

	res := (gammaM - gammaK - gammaMminK) +
		(gammaNminM - gammaSampleMinK - gammaNminMminSamplePlusK) -
		(gammaN - gammaSample - gammaNminSample)
	initialPdf := math.Exp(res)
	for i := range buffer {
		u := engine.prng.Float64()
		p := initialPdf
		f := p
		k := startK
		for u > f && float64(k) < nsample && float64(k) < n {
			k++
			p *= ((nsample - float64(k) + 1) * (m - float64(k) + 1)) / (float64(k) * (n - m - nsample + float64(k)))
			f += p
		}
		buffer[i] = float64(k)
	}

	return nil
}

func (engine *SimulatorEngine) FillExponential(buffer []float64, lambda float64) error {
	if lambda < 0 || lambda > 600 {
		return errors.New("lambda must be between 0 and 600")
	}

	for i := range buffer {
		aleatoryNumber := 1.0 - engine.prng.Float64()
		buffer[i] = -math.Log(aleatoryNumber) / lambda
	}
	return nil
}
