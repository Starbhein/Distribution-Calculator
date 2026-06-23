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

func ValidateParams(distribution string, params []float64) errorMessage {
	switch distribution {
	case "Binomial":
		if len(params) < 2 {
			return errorMessage{error: errors.New("faltan parámetros para Binomial"), index: -1}
		}
		p := params[0]
		n := params[1]
		if p <= 0 || p > 1 {
			return errorMessage{error: errors.New("p debe estar en (0, 1]"), index: 0}
		}
		if n <= 0 {
			return errorMessage{error: errors.New("n debe ser mayor que 0"), index: 1}
		}
	case "Poisson":
		if len(params) < 1 {
			return errorMessage{error: errors.New("faltan parámetros para Poisson"), index: -1}
		}
		if params[0] <= 0 {
			return errorMessage{error: errors.New("λ debe ser mayor que 0"), index: 0}
		}
	case "Hypergeométrica":
		if len(params) < 3 {
			return errorMessage{error: errors.New("faltan parámetros para Hypergeométrica"), index: -1}
		}
		N := params[0]
		M := params[1]
		n := params[2]
		if N <= 0 {
			return errorMessage{error: errors.New("N debe ser mayor que 0"), index: 0}
		}
		if M <= 0 {
			return errorMessage{error: errors.New("M debe ser mayor que 0"), index: 1}
		}
		if n <= 0 {
			return errorMessage{error: errors.New("n debe ser mayor que 0"), index: 2}
		}
		if M > N {
			return errorMessage{error: errors.New("M no puede ser mayor que N"), index: 1}
		}
		if n > N {
			return errorMessage{error: errors.New("n no puede ser mayor que N"), index: 2}
		}
	case "Normal":
		if len(params) < 2 {
			return errorMessage{error: errors.New("faltan parámetros para Normal"), index: -1}
		}
		if params[1] <= 0 {
			return errorMessage{error: errors.New("σ debe ser mayor que 0"), index: 1}
		}
	case "Exponencial (λ)":
		if len(params) < 1 {
			return errorMessage{error: errors.New("faltan parámetros para Exponencial"), index: -1}
		}
		if params[0] <= 0 {
			return errorMessage{error: errors.New("λ debe ser mayor que 0"), index: 0}
		}
	case "Exponencial (β)":
		if len(params) < 1 {
			return errorMessage{error: errors.New("faltan parámetros para Exponencial (β)"), index: -1}
		}
		if params[0] <= 0 {
			return errorMessage{error: errors.New("β debe ser mayor que 0"), index: 0}
		}
	case "Bernoulli":
		if len(params) < 1 {
			return errorMessage{error: errors.New("faltan parámetros para Bernoulli"), index: -1}
		}
		if params[0] <= 0 || params[0] > 1 {
			return errorMessage{error: errors.New("p debe estar en (0, 1]"), index: 0}
		}
	case "Geométrica":
		if len(params) < 1 {
			return errorMessage{error: errors.New("faltan parámetros para Geométrica"), index: -1}
		}
		if params[0] <= 0 || params[0] > 1 {
			return errorMessage{error: errors.New("p debe estar en (0, 1]"), index: 0}
		}
	case "Uniforme continua":
		if len(params) < 2 {
			return errorMessage{error: errors.New("faltan parámetros para Uniforme continua"), index: -1}
		}
		if params[0] >= params[1] {
			return errorMessage{error: errors.New("a debe ser menor que b"), index: 0}
		}
	default:
		return errorMessage{error: errors.New("distribución desconocida: " + distribution), index: -1}
	}
	return errorMessage{}
}

// runSimulationOnce ejecuta una simulación sincrónica y retorna datos + stats.
// Reutilizable por RunSimulationCmd y RunLLNCmd.
func runSimulationOnce(distribution string, params []float64, sampleSize int) ([]float64, stats.EmpiricalStats, error) {
	if sampleSize <= 0 {
		sampleSize = 1000
	}

	buffer := make([]float64, sampleSize)
	numGoroutines := concurrentWorkers
	if sampleSize < numGoroutines {
		numGoroutines = sampleSize
	}
	chunkSize := sampleSize / numGoroutines

	var wg sync.WaitGroup
	accumulators := make([]stats.WelfordAccumulator, numGoroutines)

	// Pre-build CDF tables once to share across workers
	var binomialCDF []float64
	var poissonCDF []float64
	var hypergeoCDF []float64
	var hypergeoErr error

	switch distribution {
	case "Binomial":
		n := int(params[1])
		p := params[0]
		variance := float64(n) * p * (1.0 - p)
		if variance <= 9.0 {
			binomialCDF = sim.BuildBinomialCDFTable(n, p)
		}
	case "Poisson":
		lambda := params[0]
		if lambda > 10.0 && lambda <= 100.0 {
			poissonCDF = sim.BuildPoissonCDFTable(lambda)
		}
	case "Hypergeométrica":
		N := params[0]
		M := params[1]
		n := params[2]
		variance := n * (M / N) * ((N - M) / N) * ((N - n) / (N - 1))
		if variance <= 9.0 {
			hypergeoCDF, _, _, hypergeoErr = sim.BuildHypergeometricCDFTable(M, n, N)
		}
	}
	if hypergeoErr != nil {
		return nil, stats.EmpiricalStats{}, hypergeoErr
	}

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
			var fillErr error

			switch distribution {
			case "Binomial":
				n := int(params[1])
				p := params[0]
				fillErr = engine.FillBinomial(subBuffer, n, p, binomialCDF)
			case "Poisson":
				fillErr = engine.FillPoisson(subBuffer, params[0], poissonCDF)
			case "Hypergeométrica":
				fillErr = engine.FillHypergeometric(subBuffer, params[1], params[2], params[0], hypergeoCDF)
			case "Normal":
				fillErr = engine.FillNormal(subBuffer, params[0], params[1])
			case "Exponencial (λ)":
				fillErr = engine.FillExponential(subBuffer, params[0])
			case "Exponencial (β)":
				fillErr = engine.FillExponential(subBuffer, 1.0/params[0])
			case "Bernoulli":
				fillErr = engine.FillBernoulli(subBuffer, params[0])
			case "Geométrica":
				fillErr = engine.FillGeometric(subBuffer, params[0])
			case "Uniforme continua":
				fillErr = engine.FillUniformContinuous(subBuffer, params[0], params[1])
			}

			if fillErr != nil {
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
