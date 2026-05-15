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
		const wantCDF = 0.433470120366708933617
		got, err := poisson.PMF(k)
		if err != nil {
			t.Error(err.Error())
		}
		gotCDF, err := poisson.CDF(k)
		if err != nil {
			t.Error(err.Error())
		}
		evalFloats(t, got, want)
		evalFloats(t, gotCDF, wantCDF)
	})
	t.Run("Poisson test with Lambda = 4, k=0", func(t *testing.T) {
		poisson, err := NewPoisson(float64(4))
		if err != nil {
			t.Error(err.Error())
		}
		const k, wantPMF = 0, 0.0183156388887341802937180212732412
		const wantAvg, wantStdDev = 4, 2
		got, err := poisson.PMF(k)
		gotAvg := poisson.Avg()
		gotStdDev := poisson.StdDev()
		if err != nil {
			t.Error(err.Error())
		}
		evalFloats(t, got, wantPMF)
		evalFloats(t, gotAvg, wantAvg)
		evalFloats(t, gotStdDev, wantStdDev)
	})
}
