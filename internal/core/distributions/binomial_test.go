package distributions

import (
	"math"
	"testing"
)

const epsilonFailure = 1e-12

func TestBinomial(t *testing.T) {
	t.Run("Binomial test with N=10 and PE= 0.1 when k = 2", func(t *testing.T) {
		const k = 2
		const want = .1937102445
		obj := Binomial{
			EP: 0.1,
			N:  10,
		}

		got, err := obj.PMF(k)
		if err != nil {
			t.Error(err.Error())
		}
		if got-want > epsilonFailure {
			t.Errorf("got %f wanted %f", got, want)
		}
	})
	t.Run("Binomial test with N=10 and PE=.1 when k < 0", func(t *testing.T) {
		obj := Binomial{
			EP: 0.1,
			N:  10,
		}
		const k = -1
		got, _ := obj.PMF(k)
		evalFloats(t, got, 0.0)
		// if err != nil {
		// 	t.Error(err.Error())
		// }
	})
	t.Run("Binomial test with N=10 and PE=.5 when k=4", func(t *testing.T) {
		obj, err := NewBinomial(10, .5)
		if err != nil {
			t.Error(err.Error())
		}
		const k = 4
		const want = 0.376953125
		const wantPMF = 0.205078125
		got, err := obj.CDF(k)
		if err != nil {
			t.Error(err.Error())
		}
		gotPMF, err := obj.PMF(k)
		evalFloats(t, gotPMF, wantPMF)
		if err != nil {
			t.Error(err.Error())
		}
		evalFloats(t, got, want)
	})
}

func evalFloats(t testing.TB, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > epsilonFailure {
		t.Errorf("got %v wanted %v, has a measurement error of: %v", got, want, math.Abs(got-want))
	}
}
