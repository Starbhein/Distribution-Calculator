package sim

import (
	"errors"
	"math"
	"math/rand/v2"
)

type SimulatorEngine struct {
	prng *rand.Rand
}

// NewSimulatorEngine uses PRNG via PCG (Permuted Congruential Generator) to simulate randomized data
func NewSimulatorEngine(seed1, seed2 uint64) *SimulatorEngine {
	source := rand.NewPCG(seed1, seed2)
	return &SimulatorEngine{prng: rand.New(source)}
}

func (engine *SimulatorEngine) FillBinomial(buffer []float64, trials int64, successProbability float64) error {
	pRatio := successProbability / float64(1-successProbability)
	for i := range buffer {
		aleatoryNumber := engine.prng.Float64()
		pdf := math.Pow((1.0 - successProbability), float64(trials))
		cdf := pdf
		k := 0

		for aleatoryNumber > cdf {
			k++
			pdf *= pRatio * (float64(int(trials)-k+1) / float64(k))
			// (successProbability * (float64(bufferSize - k + 1))) / float64(k) * (1 - successProbability)
			cdf += pdf
		}
		buffer[i] = float64(k)
	}
	return nil
}

func (engine *SimulatorEngine) FillPoisson(buffer []float64, lambda float64) error {
	if lambda < 0 || lambda > 600 {
		return errors.New("lambda must be between 0 and 600")
	}
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
