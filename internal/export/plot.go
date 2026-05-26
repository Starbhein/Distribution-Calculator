package export

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// ExportPlot generates a PNG or SVG with histogram + theoretical curve + vertical line at x.
// format: "png" or "svg"
func ExportPlot(data []float64, dist string, params []float64, markX float64, outPath string, format string) error {
	if len(data) == 0 {
		return fmt.Errorf("no hay datos para graficar")
	}

	p := plot.New()
	p.Title.Text = buildTitle(dist, params, data)
	p.X.Label.Text = "Valor"
	p.Y.Label.Text = "Densidad"
	p.Legend.Top = true

	if isDiscreteDistribution(dist) {
		return exportDiscretePlot(p, data, dist, params, markX, outPath, format)
	}
	return exportContinuousPlot(p, data, dist, params, markX, outPath, format)
}

func exportContinuousPlot(p *plot.Plot, data []float64, dist string, params []float64, markX float64, outPath string, format string) error {
	// Histograma empírico normalizado
	values := make(plotter.Values, len(data))
	for i, v := range data {
		values[i] = v
	}
	hist, err := plotter.NewHist(values, 20)
	if err != nil {
		return fmt.Errorf("error creando histograma: %w", err)
	}
	hist.Normalize(1)
	hist.FillColor = color.RGBA{R: 175, G: 238, B: 238, A: 255}
	hist.LineStyle.Color = color.RGBA{R: 80, G: 200, B: 200, A: 255}
	p.Add(hist)
	p.Legend.Add("Histograma empírico", hist)

	// Curva PDF teórica
	distParams := params[:len(params)-1]
	pdfFunc := buildPDFFunc(dist, distParams)
	var fn *plotter.Function
	if pdfFunc != nil {
		fn = plotter.NewFunction(pdfFunc)
		fn.Color = color.RGBA{R: 189, G: 147, B: 249, A: 255} // #BD93F9
		fn.Width = vg.Points(2)
		p.Add(fn)
		p.Legend.Add("PDF teórica", fn)
	}

	// Línea vertical en x = markX
	maxY := computeMaxY(p, data, pdfFunc)
	if maxY > 0 {
		line, err := plotter.NewLine(plotter.XYs{
			{X: markX, Y: 0},
			{X: markX, Y: maxY},
		})
		if err == nil {
			line.Color = color.RGBA{R: 255, G: 85, B: 85, A: 255} // #FF5555
			line.Dashes = []vg.Length{vg.Points(4), vg.Points(2)}
			line.Width = vg.Points(1.5)
			p.Add(line)
			p.Legend.Add(fmt.Sprintf("x = %.2f", markX), line)
		}
	}

	return p.Save(6*vg.Inch, 4*vg.Inch, outPath)
}

func exportDiscretePlot(p *plot.Plot, data []float64, dist string, params []float64, markX float64, outPath string, format string) error {
	// Contar frecuencias empíricas
	freq := make(map[int]int)
	for _, v := range data {
		k := int(math.Round(v))
		freq[k]++
	}

	// Empirical bars
	n := float64(len(data))
	var empBars plotter.XYs
	var theoBars plotter.XYs
	markedK := int(math.Round(markX))
	distParams := params[:len(params)-1]

	minK, maxK := math.MaxInt, math.MinInt
	for k := range freq {
		if k < minK {
			minK = k
		}
		if k > maxK {
			maxK = k
		}
	}

	for k := minK; k <= maxK; k++ {
		count := freq[k]
		empBars = append(empBars, plotter.XY{X: float64(k), Y: float64(count) / n})
		if dist != "" && distParams != nil {
			pmf := getTheoreticalPMF(dist, distParams, k)
			theoBars = append(theoBars, plotter.XY{X: float64(k), Y: pmf})
		}
	}

	// Empirical line (green, with circles)
	var empLine *plotter.Line
	var empScatter *plotter.Scatter
	if len(empBars) > 0 {
		empLine, _ = plotter.NewLine(empBars)
		if empLine != nil {
			empLine.Color = color.RGBA{R: 80, G: 250, B: 123, A: 255}
			empLine.Width = vg.Points(1.5)
			p.Add(empLine)
		}
		empScatter, _ = plotter.NewScatter(empBars)
		if empScatter != nil {
			empScatter.GlyphStyle.Color = color.RGBA{R: 80, G: 250, B: 123, A: 255}
			empScatter.GlyphStyle.Radius = vg.Points(3)
			p.Add(empScatter)
		}
		p.Legend.Add("Frecuencia empírica", empLine)
	}

	// Theoretical line (purple, with diamonds)
	var theoLine *plotter.Line
	var theoScatter *plotter.Scatter
	if len(theoBars) > 0 {
		theoLine, _ = plotter.NewLine(theoBars)
		if theoLine != nil {
			theoLine.Color = color.RGBA{R: 189, G: 147, B: 249, A: 255}
			theoLine.Width = vg.Points(2)
			p.Add(theoLine)
		}
		theoScatter, _ = plotter.NewScatter(theoBars)
		if theoScatter != nil {
			theoScatter.GlyphStyle.Color = color.RGBA{R: 189, G: 147, B: 249, A: 255}
			theoScatter.GlyphStyle.Radius = vg.Points(3)
			theoScatter.GlyphStyle.Shape = draw.RingGlyph{}
			p.Add(theoScatter)
		}
		p.Legend.Add("PMF teórica", theoLine)
	}

	// Mark bar at k=markedK with red line
	maxY := computeMaxYDiscrete(empBars, theoBars)
	if maxY > 0 {
		line, err := plotter.NewLine(plotter.XYs{
			{X: float64(markedK) - 0.5, Y: 0},
			{X: float64(markedK) - 0.5, Y: maxY},
		})
		if err == nil {
			line.Color = color.RGBA{R: 255, G: 85, B: 85, A: 255}
			line.Dashes = []vg.Length{vg.Points(4), vg.Points(2)}
			line.Width = vg.Points(1.5)
			p.Add(line)
			p.Legend.Add(fmt.Sprintf("x = %d", markedK), line)
		}
	}

	return p.Save(6*vg.Inch, 4*vg.Inch, outPath)
}

// buildTitle creates a descriptive title with distribution name, parameters, n, mean and stddev.
func buildTitle(dist string, params []float64, data []float64) string {
	n := len(data)
	mean, stddev := computeEmpiricalStats(data)
	paramStr := formatParams(dist, params)
	return fmt.Sprintf("%s\nn=%d, μ̂=%.4f, σ̂=%.4f, %s", dist, n, mean, stddev, paramStr)
}

// computeEmpiricalStats calculates mean and standard deviation from data.
func computeEmpiricalStats(data []float64) (mean, stddev float64) {
	if len(data) == 0 {
		return 0, 0
	}
	var sum float64
	for _, v := range data {
		sum += v
	}
	mean = sum / float64(len(data))

	var variance float64
	for _, v := range data {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(data) - 1)
	if len(data) == 1 {
		variance = 0
	}
	stddev = math.Sqrt(variance)
	return
}

// formatParams returns a human-readable string of distribution parameters.
func formatParams(dist string, params []float64) string {
	switch dist {
	case "Binomial":
		if len(params) >= 2 {
			return fmt.Sprintf("p=%.4f, n=%.0f", params[0], params[1])
		}
	case "Poisson":
		if len(params) >= 1 {
			return fmt.Sprintf("λ=%.4f", params[0])
		}
	case "Hypergeométrica":
		if len(params) >= 3 {
			return fmt.Sprintf("N=%.0f, M=%.0f, n=%.0f", params[0], params[1], params[2])
		}
	case "Normal":
		if len(params) >= 2 {
			return fmt.Sprintf("μ=%.4f, σ=%.4f", params[0], params[1])
		}
	case "Exponencial":
		if len(params) >= 1 {
			return fmt.Sprintf("λ=%.4f", params[0])
		}
	case "Exponencial (β)":
		if len(params) >= 1 {
			return fmt.Sprintf("β=%.4f", params[0])
		}
	case "Bernoulli":
		if len(params) >= 1 {
			return fmt.Sprintf("p=%.4f", params[0])
		}
	case "Geométrica":
		if len(params) >= 1 {
			return fmt.Sprintf("p=%.4f", params[0])
		}
	case "Uniforme continua":
		if len(params) >= 2 {
			return fmt.Sprintf("a=%.4f, b=%.4f", params[0], params[1])
		}
	}
	return ""
}

func buildPDFFunc(dist string, params []float64) func(float64) float64 {
	switch dist {
	case "Normal":
		if len(params) >= 2 {
			n, _ := distributions.NewNormal(params[0], params[1])
			if n != nil {
				return func(x float64) float64 {
					v, _ := n.PDF(x)
					return v
				}
			}
		}
	case "Exponencial":
		if len(params) >= 1 {
			el, _ := distributions.NewExponentialLambda(params[0])
			if el != nil {
				return func(x float64) float64 {
					v, _ := el.PDF(x)
					return v
				}
			}
		}
	case "Exponencial (β)":
		if len(params) >= 1 {
			eb, _ := distributions.NewExponentialBetha(params[0])
			if eb != nil {
				return func(x float64) float64 {
					v, _ := eb.PDF(x)
					return v
				}
			}
		}
	case "Uniforme continua":
		if len(params) >= 2 {
			a, b := params[0], params[1]
			return func(x float64) float64 {
				if x < a || x > b || a >= b {
					return 0
				}
				return 1.0 / (b - a)
			}
		}
	}
	return nil
}

func getTheoreticalPMF(dist string, params []float64, k int) float64 {
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

func computeMaxY(p *plot.Plot, data []float64, pdfFunc func(float64) float64) float64 {
	if pdfFunc == nil {
		return 1.0
	}
	minV, maxV := data[0], data[0]
	for _, v := range data[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	maxY := 0.0
	for i := 0; i <= 100; i++ {
		x := minV + float64(i)*(maxV-minV)/100
		if y := pdfFunc(x); y > maxY {
			maxY = y
		}
	}
	return maxY * 1.2
}

func computeMaxYDiscrete(empBars, theoBars plotter.XYs) float64 {
	maxY := 0.0
	for _, pt := range empBars {
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}
	for _, pt := range theoBars {
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}
	return maxY * 1.2
}

func isDiscreteDistribution(name string) bool {
	switch name {
	case "Binomial", "Poisson", "Hypergeométrica", "Bernoulli", "Geométrica":
		return true
	default:
		return false
	}
}

// GenerateFilename creates an auto-generated filename.
func GenerateFilename(dist string, ext string) string {
	sanitized := strings.ToLower(strings.ReplaceAll(dist, " ", "-"))
	sanitized = strings.ReplaceAll(sanitized, "(", "")
	sanitized = strings.ReplaceAll(sanitized, ")", "")
	sanitized = strings.ReplaceAll(sanitized, "β", "beta")
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("distcalc-%s-%s.%s", sanitized, timestamp, ext)
}
