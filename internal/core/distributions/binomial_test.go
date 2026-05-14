package distributions

import "testing"

const epsilonFailure = .000001

func TestBinomial(t *testing.T) {
	t.Run("Binomial test with N=10 and PE= 0.1", func(t *testing.T) {
		obj := Binomial{
			EP: 0.1,
			N:  10,
		}
		got, err := obj.PMF(2)
		if err != nil {
			t.Error(err.Error())
		}
		if got-.1937 > epsilonFailure {
			t.Errorf("got %f wanted %f", got, 0.1931)
		}
	})
}

// func evalBAvg(t testing.TB, b Binomial) {
// 	got, _ := b.Avg()
// 	want := 16
// 	if got-float64(want) > epsilonFailure {
// 		t.Errorf("want %f but got %f", float64(want), got)
// 	}
// }
