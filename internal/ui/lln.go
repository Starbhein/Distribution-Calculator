package ui

import (
	"fmt"
	"math"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Starbhein/DistCalc/internal/core/distributions/registry"
)

// LLNStep represents one step in the LLN simulation.
type LLNStep struct {
	N     int
	Mean  float64
	Delta float64 // |mean - theoreticalMean|
}

// MsgLLNDone is sent when LLN simulation completes.
type MsgLLNDone struct {
	Steps           []LLNStep
	TheoreticalMean float64
	Dist            string
	Params          []float64
}

// RunLLNCmd simulates with increasing sample sizes.
// steps: number of simulation steps (e.g., 8)
// startN: initial sample size (e.g., 10)
// growthFactor: multiply N by this each step (e.g., 2)
func RunLLNCmd(distribution string, params []float64, steps, startN, growthFactor int, theoreticalMean float64) tea.Cmd {
	return func() tea.Msg {
		results := make([]LLNStep, 0, steps)
		n := startN

		for i := 0; i < steps; i++ {
			_, empiricalStats, err := runSimulationOnce(distribution, params, n)
			if err != nil {
				return errorMessage{error: err, index: -1}
			}

			mean := empiricalStats.Avg
			delta := math.Abs(mean - theoreticalMean)

			results = append(results, LLNStep{
				N:     n,
				Mean:  mean,
				Delta: delta,
			})

			n *= growthFactor
		}

		return MsgLLNDone{
			Steps:           results,
			TheoreticalMean: theoreticalMean,
			Dist:            distribution,
			Params:          params,
		}
	}
}

// RenderLLN creates the ASCII visualization of LLN convergence.
func RenderLLN(steps []LLNStep, theoreticalMean float64, dist string, params []float64, width int) string {
	var sb strings.Builder

	// Header — param formatting is single-sourced in the registry spec
	// (design §1.3 — FormatParams replaces the triplicated formatters).
	paramStr := ""
	if spec, ok := registry.ByName(dist); ok {
		paramStr = spec.FormatParams(params)
	}
	sb.WriteString(titleh1Style.Render(fmt.Sprintf("Ley de Grandes Números — %s", dist)) + "\n")
	sb.WriteString(secondaryTextStyle().Render(paramStr) + "\n")
	sb.WriteString(fmt.Sprintf("μ teórica = %.4f\n\n", theoreticalMean))

	// Find max delta for bar scaling
	maxDelta := 0.0
	for _, s := range steps {
		if s.Delta > maxDelta {
			maxDelta = s.Delta
		}
	}
	if maxDelta == 0 {
		maxDelta = 1
	}

	barWidth := width - 50
	if barWidth < 10 {
		barWidth = 10
	}

	for _, s := range steps {
		barLen := int((s.Delta / maxDelta) * float64(barWidth))
		if barLen < 1 && s.Delta > 0 {
			barLen = 1
		}
		if s.Delta == 0 {
			barLen = 0
		}

		bar := strings.Repeat("█", barLen)
		label := fmt.Sprintf("n=%-7d", s.N)
		meanStr := fmt.Sprintf("μ̂=%8.4f", s.Mean)
		deltaStr := fmt.Sprintf("|Δ|=%.4f", s.Delta)

		sb.WriteString(fmt.Sprintf("%s |%-*s %s  %s\n", label, barWidth, bar, meanStr, deltaStr))
	}

	sb.WriteString("\n" + mutedStyle.Render("Δ = |μ̂ - μ| → 0 cuando n → ∞"))

	return sb.String()
}
