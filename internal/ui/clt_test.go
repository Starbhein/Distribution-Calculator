package ui

import (
	"math"
	"testing"
)

func TestRunCLTCmd_Bernoulli(t *testing.T) {
	// Bernoulli(p=0.5): μ=0.5, σ=0.5
	// With sampleSize=30, SE = 0.5/√30 ≈ 0.0913
	msg := RunCLTCmd("Bernoulli", []float64{0.5})()
	clt, ok := msg.(MsgCLTDone)
	if !ok {
		t.Fatalf("expected MsgCLTDone, got %T", msg)
	}

	if len(clt.Means) != cltNumSamples {
		t.Errorf("expected %d means, got %d", cltNumSamples, len(clt.Means))
	}

	// Check theoretical stats
	if math.Abs(clt.TheoreticalMean-0.5) > 1e-9 {
		t.Errorf("expected theoretical mean 0.5, got %f", clt.TheoreticalMean)
	}
	expectedSE := 0.5 / math.Sqrt(float64(cltSampleSize))
	if math.Abs(clt.TheoreticalSE-expectedSE) > 1e-9 {
		t.Errorf("expected theoretical SE %f, got %f", expectedSE, clt.TheoreticalSE)
	}

	// Empirical mean should be close to theoretical mean
	var sum float64
	for _, v := range clt.Means {
		sum += v
	}
	empMean := sum / float64(len(clt.Means))
	if math.Abs(empMean-clt.TheoreticalMean) > 0.02 {
		t.Errorf("empirical mean %f too far from theoretical %f", empMean, clt.TheoreticalMean)
	}
}

func TestRunCLTCmd_Normal(t *testing.T) {
	// Normal(μ=10, σ=2): SE = 2/√30 ≈ 0.365
	msg := RunCLTCmd("Normal", []float64{10, 2})()
	clt, ok := msg.(MsgCLTDone)
	if !ok {
		t.Fatalf("expected MsgCLTDone, got %T", msg)
	}

	if len(clt.Means) != cltNumSamples {
		t.Errorf("expected %d means, got %d", cltNumSamples, len(clt.Means))
	}

	if math.Abs(clt.TheoreticalMean-10.0) > 1e-9 {
		t.Errorf("expected theoretical mean 10, got %f", clt.TheoreticalMean)
	}

	expectedSE := 2.0 / math.Sqrt(float64(cltSampleSize))
	if math.Abs(clt.TheoreticalSE-expectedSE) > 1e-9 {
		t.Errorf("expected theoretical SE %f, got %f", expectedSE, clt.TheoreticalSE)
	}

	var sum float64
	for _, v := range clt.Means {
		sum += v
	}
	empMean := sum / float64(len(clt.Means))
	if math.Abs(empMean-10.0) > 0.05 {
		t.Errorf("empirical mean %f too far from 10", empMean)
	}
}

func TestRenderCLT_NotEmpty(t *testing.T) {
	means := make([]float64, 100)
	for i := range means {
		means[i] = 5.0 + float64(i)*0.01
	}
	rendered := RenderCLT(means, "Normal", []float64{5, 1}, 5.0, 0.2, 80, 40)
	if rendered == "" || rendered == "Sin datos para graficar" {
		t.Error("RenderCLT produced empty output")
	}
}
