package ui

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"github.com/Starbhein/DistCalc/internal/core/distributions/registry"
	"github.com/Starbhein/DistCalc/internal/core/distributions/sim"
	"github.com/Starbhein/DistCalc/internal/core/stats"
)

// CLT constants — tuned for visual clarity and speed.
const (
	cltSampleSize = 30   // size of each sample
	cltNumSamples = 5000 // number of sample means to compute
)

// MsgCLTDistributionSelected is sent when the user picks a distribution from the CLT sub-menu.
type MsgCLTDistributionSelected struct {
	Distribution string
}

// MsgCLTDone is sent when CLT simulation completes.
type MsgCLTDone struct {
	Means           []float64
	Dist            string
	Params          []float64
	TheoreticalMean float64
	TheoreticalSE   float64 // standard error = σ/√n
}

// RunCLTCmd simulates the Central Limit Theorem.
// It draws `cltNumSamples` samples of size `cltSampleSize`, computes the mean of each,
// and returns the distribution of sample means.
func RunCLTCmd(distribution string, params []float64) tea.Cmd {
	return func() tea.Msg {
		theo, err := ComputeTheoreticalStats(distribution, params)
		if err != nil {
			return errorMessage{error: err, index: -1}
		}

		// Fill dispatch goes through the registry sampler (design §1.3):
		// ByName + NewSampler + one Prebuild per run + Fill per worker.
		spec, ok := registry.ByName(distribution)
		if !ok {
			return errorMessage{error: errors.New("distribución desconocida: " + distribution), index: -1}
		}
		sampler, err := spec.NewSampler(params)
		if err != nil {
			return errorMessage{error: err, index: -1}
		}
		if err := sampler.Prebuild(); err != nil {
			return errorMessage{error: err, index: -1}
		}

		means := make([]float64, cltNumSamples)

		numGoroutines := concurrentWorkers
		if cltNumSamples < numGoroutines {
			numGoroutines = cltNumSamples
		}

		var wg sync.WaitGroup

		for g := 0; g < numGoroutines; g++ {
			start := g * cltNumSamples / numGoroutines
			end := (g + 1) * cltNumSamples / numGoroutines
			count := end - start
			if count <= 0 {
				continue
			}

			wg.Add(1)
			go func(id, startIdx, sampleCount int) {
				defer wg.Done()

				engine := sim.NewSimulatorEngine(uint64(id+1), 99)
				buffer := make([]float64, sampleCount*cltSampleSize)

				// Fill the entire batch in one call for cache efficiency
				if err := sampler.Fill(engine, buffer); err != nil {
					return
				}

				// Compute sample means
				for i := 0; i < sampleCount; i++ {
					offset := i * cltSampleSize
					var sum float64
					for j := 0; j < cltSampleSize; j++ {
						sum += buffer[offset+j]
					}
					means[startIdx+i] = sum / float64(cltSampleSize)
				}
			}(g, start, count)
		}

		wg.Wait()

		return MsgCLTDone{
			Means:           means,
			Dist:            distribution,
			Params:          params,
			TheoreticalMean: theo.Avg,
			TheoreticalSE:   theo.StdDev / math.Sqrt(float64(cltSampleSize)),
		}
	}
}

// RenderCLT creates the ASCII visualization of the CLT.
func RenderCLT(means []float64, dist string, params []float64, theoMean, theoSE float64, width, height int) string {
	if len(means) == 0 {
		return "Sin datos para graficar"
	}

	var sb strings.Builder

	// Header — param formatting is single-sourced in the registry spec
	// (design §1.3 — FormatParams replaces the triplicated formatters).
	paramStr := ""
	if spec, ok := registry.ByName(dist); ok {
		paramStr = spec.FormatParams(params)
	}
	sb.WriteString(titleh1Style.Render(fmt.Sprintf("Teorema del Límite Central — %s", dist)) + "\n")
	sb.WriteString(secondaryTextStyle().Render(paramStr) + "\n")
	sb.WriteString(fmt.Sprintf("Muestras=%d  Tamaño=%d\n", cltNumSamples, cltSampleSize))
	sb.WriteString(fmt.Sprintf("μ teórica = %.4f  SE teórico = %.4f\n\n", theoMean, theoSE))

	// Empirical stats of the means
	empStats := stats.AnalyzeBuffer(means)
	sb.WriteString(fmt.Sprintf(
		"Media de medias = %.4f  Desv. estándar = %.4f\n\n",
		empStats.Avg, math.Sqrt(empStats.Variance),
	))

	// Histogram of sample means + theoretical normal overlay
	chartW := (width * 70) / 100
	if chartW < 20 {
		chartW = 20
	}
	maxChartLines := height - 22
	if maxChartLines < 6 {
		maxChartLines = 6
	}

	hist := renderCLTHistogram(means, theoMean, theoSE, chartW, maxChartLines)
	sb.WriteString(hist)

	sb.WriteString("\n" + mutedStyle.Render("Las medias muestrales se distribuyen ~N(μ, σ/√n)"))

	return sb.String()
}

// renderCLTHistogram draws a continuous histogram of sample means with a theoretical normal curve.
func renderCLTHistogram(data []float64, theoMean, theoSE float64, contentWidth, maxLines int) string {
	minV, maxV := minMax(data)
	if minV == maxV {
		return "Todos los valores son iguales"
	}

	binCount := 16
	if len(data) < binCount {
		binCount = len(data)
	}
	if binCount < 2 {
		binCount = 2
	}

	binWidth := (maxV - minV) / float64(binCount)
	if binWidth == 0 {
		binWidth = 1
	}

	bins := make([]int, binCount)
	for _, v := range data {
		idx := int((v - minV) / binWidth)
		if idx >= binCount {
			idx = binCount - 1
		}
		if idx < 0 {
			idx = 0
		}
		bins[idx]++
	}

	// Normalize to density
	n := float64(len(data))
	densities := make([]float64, binCount)
	maxDensity := 0.0
	for i, c := range bins {
		densities[i] = float64(c) / (n * binWidth)
		if densities[i] > maxDensity {
			maxDensity = densities[i]
		}
	}

	// Theoretical normal PDF at bin centers
	pdfValues := make([]float64, binCount)
	maxPDF := 0.0
	for i := 0; i < binCount; i++ {
		center := minV + (float64(i)+0.5)*binWidth
		pdfValues[i] = normalPDF(center, theoMean, theoSE)
		if pdfValues[i] > maxPDF {
			maxPDF = pdfValues[i]
		}
	}

	maxY := maxDensity
	if maxPDF > maxY {
		maxY = maxPDF
	}
	if maxY == 0 {
		maxY = 1
	}
	maxY *= 1.05

	labelWidth := 12
	barWidth := contentWidth - labelWidth - 24
	if barWidth < 5 {
		barWidth = 5
	}

	var sb strings.Builder
	for i, count := range bins {
		lo := minV + float64(i)*binWidth
		hi := minV + float64(i+1)*binWidth
		density := densities[i]

		barLen := 0
		if maxY > 0 {
			barLen = int((density / maxY) * float64(barWidth))
		}
		if barLen < 1 && count > 0 {
			barLen = 1
		}

		bar := strings.Repeat("█", barLen)
		label := fmt.Sprintf("%.2f-%.2f", lo, hi)
		if len(label) > labelWidth {
			label = fmt.Sprintf("%.1f-%.1f", lo, hi)
		}

		sb.WriteString(fmt.Sprintf("%12s |%s dens:%.3f pdf:%.3f\n", label, bar, density, pdfValues[i]))
	}

	return sb.String()
}

// normalPDF evaluates the PDF of N(mean, sd) at x.
func normalPDF(x, mean, sd float64) float64 {
	if sd == 0 {
		return 0
	}
	z := (x - mean) / sd
	return math.Exp(-0.5*z*z) / (sd * math.Sqrt(2*math.Pi))
}
