package ui

import (
	"errors"
	"strconv"
	"sync"

	tea "charm.land/bubbletea/v2"
	"github.com/Starbhein/DistCalc/internal/core/distributions/sim"
	"github.com/Starbhein/DistCalc/internal/core/stats"
)

const concurrentWorkers = 4

type errorMessage struct {
	error error
	index int
}
type MsgSimulationSuccess struct {
	Stats stats.EmpiricalStats
	Data  []float64
}
type PoissonTask struct {
	Lambda float64
}

func (pt PoissonTask) FillBuffer(buffer []float64) error {
	return nil
}

func Parser(parameters []string) ([]float64, errorMessage) {
	var err error
	res := make([]float64, len(parameters))
	for i, v := range parameters {
		res[i], err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, errorMessage{error: errors.New("ingrese un numero valido"), index: i}
		}
	}
	return res, errorMessage{}
}

func LaunchSimulationCmd(sampleSize int, task sim.EmpiricalSimulator) tea.Cmd {
	return func() tea.Msg {
		buffer := make([]float64, sampleSize)
		numGoroutines := concurrentWorkers
		chunkSize := sampleSize / numGoroutines
		var wg sync.WaitGroup
		accumulators := make([]stats.WelfordAccumulator, numGoroutines)
		for g := 0; g < numGoroutines; g++ {
			wg.Add(1)
			start := g * chunkSize
			end := start + chunkSize
			if g == numGoroutines-1 {
				end = sampleSize
			}

			go func(id int, subBuffer []float64) {
				defer wg.Done()

				simulator := sim.NewSimulatorEngine(uint64(id+1), 42)
				simulator.Fill //???
				for _, val := range subBuffer {
					accumulators[id].Update(val)
				}
			}(g, buffer[start:end])
		}
	}
}
