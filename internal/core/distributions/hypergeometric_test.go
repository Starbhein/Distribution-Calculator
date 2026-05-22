package distributions

import "testing"

func TestHypergeometric(t *testing.T) {
	const successQuantity = 3
	const populationQuantity = 12
	const sampleQuantity = 4
	obj, err := NewHypergeometric(successQuantity, populationQuantity, sampleQuantity)
	t.Run("Hypergeometric test variance", func(t *testing.T) {
		if err != nil {
			t.Error(err.Error())
		}
		const wantVariance = float64(6) / float64(11)
		gotVariance := obj.Variance()
		evalFloats(t, gotVariance, wantVariance)
	})
	t.Run("Hypergeometric test Avg", func(t *testing.T) {
		if err != nil {
			t.Error(err.Error())
		}
		const wantAvg = 1

		gotAvg := obj.Avg()
		if err != nil {
			t.Error(err.Error())
		}
		evalFloats(t, gotAvg, wantAvg)
	})
	t.Run("Hypergeometric test pmf", func(t *testing.T) {
		if err != nil {
			t.Error(err.Error())
		}
		const wantPMF = float64(14) / float64(55)

		gotPMF, err := obj.PMF(0)
		if err != nil {
			t.Error(err.Error())
		}
		evalFloats(t, gotPMF, wantPMF)
	})

	t.Run("Hypergeometric test cdf", func(t *testing.T) {
		if err != nil {
			t.Error(err.Error())
		}
		const wantPMF = float64(54) / float64(55)

		gotPMF, err := obj.CDF(2)
		if err != nil {
			t.Error(err.Error())
		}
		evalFloats(t, gotPMF, wantPMF)
	})
	t.Run("Hypergeometric the sum of pmf from 0 to k=N has to return 1 ", func(t *testing.T) {
		const want = 1
		var sum float64
		for i := 0; i <= obj.M; i++ {
			v, _ := obj.PMF(i)
			sum += v
		}
		evalFloats(t, sum, want)
	})
}
