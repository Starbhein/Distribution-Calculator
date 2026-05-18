package sim

import (
	"fmt"
	"testing"

	"github.com/Starbhein/DistCalc/internal/core/stats"
)

const epsilon = .02

func TestFillBinomial(t *testing.T) {
	t.Run("simulating a binomial distribution with ep = .5 and N=50", func(t *testing.T) {
		testingSlice := make([]float64, 1000000)
		eng := NewSimulatorEngine(42, 42)
		err := eng.FillBinomial(testingSlice, 50, 0.5)
		if err != nil {
			t.Error(err)
		}

		value := stats.AnalyzeBuffer(testingSlice)

		fmt.Printf("%v \t %v \t %v ", value.Avg, value.Count, value.Variance)
		if value.Avg-25.0 > epsilon {
			t.Errorf("avg: %v got wanted %v, error of %v", value.Avg, 25.0, value.Avg-15.0)
		}

		if value.Variance-12.5 > epsilon {
			t.Errorf("variance: %v got wanted %v, error of %v", value.Variance, 12.5, value.Avg-12.5)
		}
	})
}

func TestFillPoisson(t *testing.T) {
	t.Run("simulating a poisson distribution with the seed 42,42", func(t *testing.T) {
		testingSlice := make([]float64, 1000000)
		eng := NewSimulatorEngine(42, 42)
		err := eng.FillPoisson(testingSlice, 15.0)
		if err != nil {
			t.Error(err)
		}
		value := stats.AnalyzeBuffer(testingSlice)
		fmt.Printf("%v \t %v \t %v ", value.Avg, value.Count, value.Variance)
		if value.Avg-15.0 > epsilon {
			t.Errorf("%v got wanted %v, error of %v", value.Avg, 15.0, value.Avg-15.0)
		}
	})
}
