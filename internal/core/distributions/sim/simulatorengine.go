package sim

import (
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

func (engine *SimulatorEngine) FillPoisson(buffer []float64, lambda float64) error {
	// Knuth's random generator poisson algorithm
	L := math.Exp(-lambda)
	for i := range buffer {
		k := 0
		p := float64(1)
		for p > L {
			k++
			p *= engine.prng.Float64()
		}
		buffer[i] = float64(k - 1)
	}
}
