package distributions

import "testing"

const epsilonFailure = .000001

func TestBinomial(t *testing.T) {
	t.Run("Binomial test with N=20 and PE= 0.8", func(t *testing.T) {
		obj := Binomial{
			EP: 0.8,
			N:  20,
		}
	})
}

func evalBAvg(t testing.TB, b Binomial) {
	got, _ := b.Avg()
	want := 16
	if got-float64(want) > epsilonFailure {
		t.Errorf("want %f but got %f", float64(want), got)
	}
}
