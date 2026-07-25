package ui

import (
	"errors"
	"strconv"
	"sync"

	tea "charm.land/bubbletea/v2"
	"github.com/Starbhein/DistCalc/internal/core/distributions/registry"
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

// validateParams runs THE single validation layer — the registry
// Spec.Validate (design §7, spec §3) — and maps the result to the
// (index, error) form-highlight shape. ByName is the only name-based
// dispatch left; the per-distribution rules live in the specs.
func validateParams(distribution string, params []float64) errorMessage {
	spec, ok := registry.ByName(distribution)
	if !ok {
		return errorMessage{error: errors.New("distribución desconocida: " + distribution), index: -1}
	}
	if idx, err := spec.Validate(params); err != nil {
		return errorMessage{error: err, index: idx}
	}
	return errorMessage{}
}

// runSimulationOnce ejecuta una simulación sincrónica y retorna datos + stats.
// Reutilizable por RunSimulationCmd y RunLLNCmd.
// Fill dispatch goes through the registry sampler (design §1.3): ByName +
// NewSampler + one Prebuild per run + Fill per worker.
func runSimulationOnce(distribution string, params []float64, sampleSize int) ([]float64, stats.EmpiricalStats, error) {
	if sampleSize <= 0 {
		sampleSize = 1000
	}

	spec, ok := registry.ByName(distribution)
	if !ok {
		return nil, stats.EmpiricalStats{}, errors.New("distribución desconocida: " + distribution)
	}
	sampler, err := spec.NewSampler(params)
	if err != nil {
		return nil, stats.EmpiricalStats{}, err
	}
	// Pre-build shared CDF tables once, shared across workers (no-op for
	// table-less specs).
	if err := sampler.Prebuild(); err != nil {
		return nil, stats.EmpiricalStats{}, err
	}

	buffer := make([]float64, sampleSize)
	numGoroutines := concurrentWorkers
	if sampleSize < numGoroutines {
		numGoroutines = sampleSize
	}
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

			engine := sim.NewSimulatorEngine(uint64(id+1), 42)
			if err := sampler.Fill(engine, subBuffer); err != nil {
				return
			}

			for _, val := range subBuffer {
				accumulators[id].Update(val)
			}
		}(g, buffer[start:end])
	}

	wg.Wait()

	merged := accumulators[0]
	for i := 1; i < len(accumulators); i++ {
		merged = stats.MergeWelford(merged, accumulators[i])
	}

	variance := 0.0
	if merged.Count > 1 {
		variance = merged.M2 / (merged.Count - 1)
	}

	empiricalStats := stats.EmpiricalStats{
		Count:    int64(merged.Count),
		Avg:      merged.Avg,
		Variance: variance,
	}

	return buffer, empiricalStats, nil
}

func RunSimulationCmd(distribution string, params []float64, sampleSize int) tea.Cmd {
	return func() tea.Msg {
		buffer, empiricalStats, err := runSimulationOnce(distribution, params, sampleSize)
		if err != nil {
			return errorMessage{error: err, index: -1}
		}
		return MsgSimulationSuccess{
			Stats: empiricalStats,
			Data:  buffer,
		}
	}
}
