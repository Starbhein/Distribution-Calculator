package sim

import (
	"fmt"
	"math"
	"testing"

	"github.com/Starbhein/DistCalc/internal/core/stats"
)

const (
	epsilon    = .1
	bufferSize = 1e7
	seed       = 42
)

func TestFillBinomial(t *testing.T) {
	t.Run("simulating a binomial distribution with ep = .5 and N=50", func(t *testing.T) {
		testingSlice := make([]float64, bufferSize)
		eng := NewSimulatorEngine(seed, seed)
		err := eng.FillBinomial(testingSlice, 50, 0.5, nil)
		if err != nil {
			t.Error(err)
		}
		const wantedAvg = 25.0
		const wantedVariance = 12.5
		result := stats.AnalyzeBuffer(testingSlice)
		fmt.Printf("%v\n%v\n%v", result.Variance, result.Avg, result.Count)
		evalEmpiricalStats(t, result, wantedAvg, wantedVariance)
	})
}

func TestFillPoisson(t *testing.T) {
	t.Run("simulating a poisson distribution with the seed seed,seed", func(t *testing.T) {
		testingSlice := make([]float64, bufferSize)
		eng := NewSimulatorEngine(seed, seed)
		err := eng.FillPoisson(testingSlice, 15.0, nil)
		if err != nil {
			t.Error(err)
		}
		const wantedAvg = 15.0
		const wantedVariance = 15.0
		result := stats.AnalyzeBuffer(testingSlice)
		evalEmpiricalStats(t, result, wantedAvg, wantedVariance)
	})
}

func TestFillHypergeometric(t *testing.T) {
	t.Run("simulating a hypergeometric distribution with N=50, m=20,n=10", func(t *testing.T) {
		testingSlice := make([]float64, bufferSize)
		eng := NewSimulatorEngine(seed, seed)
		err := eng.FillHypergeometric(testingSlice, 20, 10, 50, nil)
		if err != nil {
			t.Error(err)
		}
		const wantAvg = 10.0 * (20.0 / 50.0)
		const wantVariance = 1.959
		result := stats.AnalyzeBuffer(testingSlice)
		evalEmpiricalStats(t, result, wantAvg, wantVariance)
	})
}

func TestFillExponential(t *testing.T) {
	t.Run("simulating an exponential distribution with the seed seed,seed", func(t *testing.T) {
		testingSlice := make([]float64, bufferSize)
		eng := NewSimulatorEngine(seed, seed)
		const wantedAvg = 5
		const wantedVariance = 25
		err := eng.FillExponential(testingSlice, .2)
		if err != nil {
			t.Error(err)
		}

		result := stats.AnalyzeBuffer(testingSlice)
		evalEmpiricalStats(t, result, wantedAvg, wantedVariance)
	})
}

func evalEmpiricalStats(t testing.TB, result stats.EmpiricalStats, avg, variance float64) {
	t.Helper()
	if math.Abs(result.Avg-avg) > epsilon {
		t.Errorf("%v got wanted %v, error of %v", result.Avg, avg, result.Avg-avg)
	}
	if math.Abs(result.Variance-variance) > epsilon {
		t.Errorf("%v got wanted %v, error of %v", result.Variance, variance, result.Variance-variance)
	}
}
