package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
)

// RenderHistogram genera un histograma ASCII adaptado al ancho disponible.
// Para discretas usa conteo por valor entero; para continuas usa bins.
// markValue resalta el bin o barra que corresponde al valor x del usuario.
func RenderHistogram(data []float64, isDiscrete bool, contentWidth int, markValue float64) string {
	if len(data) == 0 {
		return "Sin datos para graficar"
	}

	if isDiscrete {
		return renderDiscrete(data, contentWidth, markValue)
	}
	return renderContinuous(data, contentWidth, markValue)
}

// RenderDiscreteHistogram muestra histograma empírico con PMF teórica al lado.
func RenderDiscreteHistogram(data []float64, dist string, params []float64, markValue float64, contentWidth int, maxLines int) string {
	if len(data) == 0 {
		return "Sin datos para graficar"
	}
	return renderDiscreteWithTheory(data, dist, params, markValue, contentWidth, maxLines)
}

func renderDiscrete(data []float64, contentWidth int, markValue float64) string {
	return renderDiscreteWithTheory(data, "", nil, markValue, contentWidth, 0)
}

// renderDiscreteWithTheory muestra histograma empírico + PMF teórica para discretas.
func renderDiscreteWithTheory(data []float64, dist string, params []float64, markValue float64, contentWidth int, maxLines int) string {
	freq := make(map[int]int)
	for _, v := range data {
		k := int(math.Round(v))
		freq[k]++
	}

	keys := make([]int, 0, len(freq))
	for k := range freq {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	maxFreq := 0
	for _, c := range freq {
		if c > maxFreq {
			maxFreq = c
		}
	}

	labelWidth := 6
	valWidth := 20 // espacio para "teo:0.1234 emp:0.1234"
	barWidth := contentWidth - labelWidth - valWidth - 2
	if barWidth < 5 {
		barWidth = 5
	}

	markedK := int(math.Round(markValue))

	selectedKeys := keys
	truncated := 0
	if maxLines > 0 && len(keys) > maxLines {
		type pair struct {
			k     int
			count int
		}
		pairs := make([]pair, 0, len(keys))
		for _, k := range keys {
			pairs = append(pairs, pair{k, freq[k]})
		}
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].count > pairs[j].count
		})

		selected := make(map[int]bool, maxLines)
		for i := 0; i < maxLines-1 && i < len(pairs); i++ {
			selected[pairs[i].k] = true
		}
		if !selected[markedK] {
			selected[markedK] = true
			if len(selected) > maxLines {
				for i := len(pairs) - 1; i >= 0; i-- {
					if pairs[i].k != markedK && selected[pairs[i].k] {
						delete(selected, pairs[i].k)
						break
					}
				}
			}
		}

		selectedKeys = make([]int, 0, len(selected))
		for k := range selected {
			selectedKeys = append(selectedKeys, k)
		}
		sort.Ints(selectedKeys)
		truncated = len(keys) - len(selectedKeys)
	}

	var sb strings.Builder
	for _, k := range selectedKeys {
		count := freq[k]
		barLen := 0
		if maxFreq > 0 {
			barLen = count * barWidth / maxFreq
		}
		if barLen < 1 && count > 0 {
			barLen = 1
		}

		char := "█"
		marker := ""
		if k == markedK {
			char = "▓"
			marker = " ← tu x"
		}

		bar := strings.Repeat(char, barLen)
		pct := float64(count) * 100.0 / float64(len(data))

		// PMF teórica si tenemos distribución
		var theoryStr string
		if dist != "" && params != nil {
			theo := getTheoreticalPMFChart(dist, params, k)
			// empirico como proporción (no porcentaje) para comparar con teórica
			empProp := float64(count) / float64(len(data))
			theoryStr = fmt.Sprintf(" teo:%.4f emp:%.4f", theo, empProp)
		} else {
			theoryStr = fmt.Sprintf(" %.1f%%", pct)
		}

		sb.WriteString(fmt.Sprintf("%4d |%s%s%s\n", k, bar, theoryStr, marker))
	}
	if truncated > 0 {
		sb.WriteString(mutedStyle.Render(fmt.Sprintf("... y %d valores más\n", truncated)))
	}

	return sb.String()
}

func getTheoreticalPMFChart(dist string, params []float64, k int) float64 {
	switch dist {
	case "Binomial":
		if len(params) >= 2 {
			b, _ := distributions.NewBinomial(int(params[1]), params[0])
			if b != nil {
				v, _ := b.PMF(k)
				return v
			}
		}
	case "Poisson":
		if len(params) >= 1 {
			p, _ := distributions.NewPoisson(params[0])
			if p != nil {
				v, _ := p.PMF(k)
				return v
			}
		}
	case "Hypergeométrica":
		if len(params) >= 3 {
			h, _ := distributions.NewHypergeometric(int(params[1]), int(params[0]), int(params[2]))
			if h != nil {
				v, _ := h.PMF(k)
				return v
			}
		}
	case "Bernoulli":
		if len(params) >= 1 {
			b, _ := distributions.NewBernoulli(params[0])
			if b != nil {
				v, _ := b.PMF(k)
				return v
			}
		}
	case "Geométrica":
		if len(params) >= 1 {
			g, _ := distributions.NewGeometric(params[0])
			if g != nil {
				v, _ := g.PMF(k)
				return v
			}
		}
	}
	return 0
}

// RenderContinuousHistogram muestra histograma empírico con PDF teórica al lado.
func RenderContinuousHistogram(data []float64, dist string, params []float64, markValue float64, contentWidth int, maxLines int) string {
	if len(data) == 0 {
		return "Sin datos para graficar"
	}
	return renderContinuousWithTheory(data, dist, params, markValue, contentWidth)
}

func renderContinuous(data []float64, contentWidth int, markValue float64) string {
	return renderContinuousWithTheory(data, "", nil, markValue, contentWidth)
}

// renderContinuousWithTheory muestra histograma + densidad empírica + PDF teórica.
func renderContinuousWithTheory(data []float64, dist string, params []float64, markValue float64, contentWidth int) string {
	minV, maxV := minMax(data)
	if minV == maxV {
		return "Todos los valores son iguales"
	}

	binCount := 12
	if len(data) < binCount {
		binCount = len(data)
	}
	if binCount < 2 {
		binCount = 2
	}

	bins := make([]int, binCount)
	binWidth := (maxV - minV) / float64(binCount)
	if binWidth == 0 {
		binWidth = 1
	}

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

	// Normalizar a densidad para comparar con PDF
	n := float64(len(data))
	densities := make([]float64, binCount)
	maxDensity := 0.0
	for i, c := range bins {
		densities[i] = float64(c) / (n * binWidth)
		if densities[i] > maxDensity {
			maxDensity = densities[i]
		}
	}

	// PDF teórica evaluada en centros de bins
	theoryAvailable := dist != "" && params != nil
	pdfValues := make([]float64, binCount)
	maxPDF := 0.0
	if theoryAvailable {
		distParams := params[:len(params)-1]
		for i := 0; i < binCount; i++ {
			center := minV + (float64(i)+0.5)*binWidth
			pdfValues[i] = getTheoreticalPDFChart(dist, distParams, center)
			if pdfValues[i] > maxPDF {
				maxPDF = pdfValues[i]
			}
		}
	}

	maxY := maxDensity
	if theoryAvailable && maxPDF > maxY {
		maxY = maxPDF
	}
	if maxY == 0 {
		maxY = 1
	}
	maxY *= 1.05

	labelWidth := 12
	barWidth := contentWidth - labelWidth - 22
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

		char := "█"
		marker := ""
		if markValue >= lo && markValue < hi {
			char = "▓"
			marker = " ← tu x"
		}
		if i == binCount-1 && markValue >= lo && markValue <= hi {
			char = "▓"
			marker = " ← tu x"
		}

		bar := strings.Repeat(char, barLen)
		label := fmt.Sprintf("%.1f-%.1f", lo, hi)
		if len(label) > labelWidth {
			label = fmt.Sprintf("%.0f-%.0f", lo, hi)
		}

		infoStr := ""
		if theoryAvailable {
			infoStr = fmt.Sprintf(" dens:%.3f pdf:%.3f", density, pdfValues[i])
		} else {
			infoStr = fmt.Sprintf(" %.1f%%", float64(count)*100.0/n)
		}

		sb.WriteString(fmt.Sprintf("%12s |%s%s%s\n", label, bar, infoStr, marker))
	}

	return sb.String()
}

func getTheoreticalPDFChart(dist string, params []float64, x float64) float64 {
	switch dist {
	case "Normal":
		if len(params) >= 2 {
			n, _ := distributions.NewNormal(params[0], params[1])
			if n != nil {
				v, _ := n.PDF(x)
				return v
			}
		}
	case "Exponencial (λ)":
		if len(params) >= 1 {
			el, _ := distributions.NewExponentialLambda(params[0])
			if el != nil {
				v, _ := el.PDF(x)
				return v
			}
		}
	case "Exponencial (β)":
		if len(params) >= 1 {
			eb, _ := distributions.NewExponentialBeta(params[0])
			if eb != nil {
				v, _ := eb.PDF(x)
				return v
			}
		}
	case "Uniforme continua":
		if len(params) >= 2 {
			a, b := params[0], params[1]
			if x >= a && x <= b && a < b {
				return 1.0 / (b - a)
			}
		}
	}
	return 0
}

func minMax(data []float64) (min, max float64) {
	min = data[0]
	max = data[0]
	for _, v := range data[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return
}

func isDiscreteDistribution(name string) bool {
	switch name {
	case "Binomial", "Poisson", "Hypergeométrica", "Bernoulli", "Geométrica":
		return true
	default:
		return false
	}
}
