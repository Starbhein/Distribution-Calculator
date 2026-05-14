package distributions

import (
	"testing"
)

func TestPoisson(t *testing.T) {
	t.Run("Poisson test with Lambda = 4", func(t *testing.T) {
		poisson, err := NewPoisson(float64(4))
		if err != nil {
			t.Error(err.Error())
		}
		const k, want = 3, 0.19536681481316
		got, err := poisson.PMF(k)
		if err != nil {
			t.Error(err.Error())
		}
		evalFloats(t, got, want)
	})
}
