package ui

import (
	"fmt"
	"math"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Starbhein/DistCalc/internal/core/stats"
	"github.com/Starbhein/DistCalc/internal/export"
)

type sessionState int

const (
	stateMenu sessionState = iota
	stateForm
	stateLoading
	stateResults
)

type MainModel struct {
	styles             styles
	darkBG             bool
	state              sessionState
	menu               MenuModel
	form               FormModel
	width, height      int
	spinner            spinner.Model
	empiricalStats     stats.EmpiricalStats
	chartBuffer        []float64
	activeDistribution string
	distParams         []float64
	chartView          string
	exportMsg          string
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			switch m.state {
			case stateForm, stateResults:
				m.state = stateMenu
				m.chartBuffer = nil
				m.distParams = nil
				m.chartView = ""
				m.exportMsg = ""
				return m, nil
			}
		case "e":
			if m.state == stateResults {
				filename := export.GenerateFilename(m.activeDistribution, "png")
				markValue := m.distParams[len(m.distParams)-1]
				if err := export.ExportPlot(m.chartBuffer, m.activeDistribution, m.distParams, markValue, filename, "png"); err != nil {
					m.exportMsg = fmt.Sprintf("Error PNG: %v", err)
				} else {
					m.exportMsg = fmt.Sprintf("✓ PNG guardado: %s", filename)
				}
				return m, nil
			}
		case "s":
			if m.state == stateResults {
				filename := export.GenerateFilename(m.activeDistribution, "svg")
				markValue := m.distParams[len(m.distParams)-1]
				if err := export.ExportPlot(m.chartBuffer, m.activeDistribution, m.distParams, markValue, filename, "svg"); err != nil {
					m.exportMsg = fmt.Sprintf("Error SVG: %v", err)
				} else {
					m.exportMsg = fmt.Sprintf("✓ SVG guardado: %s", filename)
				}
				return m, nil
			}
		case "c":
			if m.state == stateResults {
				filename := export.GenerateFilename(m.activeDistribution, "csv")
				if err := export.ExportCSV(m.chartBuffer, filename); err != nil {
					m.exportMsg = fmt.Sprintf("Error CSV: %v", err)
				} else {
					m.exportMsg = fmt.Sprintf("✓ CSV guardado: %s", filename)
				}
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.menu.menu.SetSize(m.width, m.height)
		return m, nil
	case MsgSelectedDistribution:
		m.state = stateForm
		m.form.BuildInputs(msg.Distribution)
		return m, nil
	case MsgForm:
		parsed, errM := Parser(msg.Parameters)
		if errM.error != nil {
			return m, func() tea.Msg {
				return errM
			}
		}
		if len(parsed) == 0 {
			return m, nil
		}
		sampleSize := int(parsed[len(parsed)-1])
		m.distParams = parsed[:len(parsed)-1]
		m.activeDistribution = m.form.activeDistribution
		m.state = stateLoading
		return m, RunSimulationCmd(m.activeDistribution, m.distParams, sampleSize)
	case MsgSimulationSuccess:
		m.state = stateResults
		m.empiricalStats = msg.Stats
		m.chartBuffer = msg.Data
		m.exportMsg = ""
		// Pre-render chart
		if len(m.distParams) > 0 {
			markValue := m.distParams[len(m.distParams)-1]
			chartW := (m.width * 55) / 100
			chartH := (m.height * 35) / 100
			if chartW < 20 {
				chartW = 20
			}
			if chartH < 8 {
				chartH = 8
			}
			// Histograma ASCII para ambos tipos
			rightWidth := (m.width*70)/100 - 4
			contentWidth := rightWidth - 8
			if contentWidth < 20 {
				contentWidth = 20
			}
			if isDiscreteDistribution(m.activeDistribution) {
				m.chartView = RenderDiscreteHistogram(m.chartBuffer, m.activeDistribution, m.distParams, markValue, contentWidth)
			} else {
				m.chartView = RenderContinuousHistogram(m.chartBuffer, m.activeDistribution, m.distParams, markValue, contentWidth)
			}
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	switch m.state {
	case stateMenu:
		updateModel, cmd := m.menu.Update(msg)
		m.menu = updateModel
		return m, cmd
	case stateForm:
		m.form, cmd = m.form.Update(msg)
		return m, cmd
	}
	return m, cmd
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
		m.spinner.Tick,
	)
}

func (m MainModel) View() tea.View {
	if m.state == stateMenu {
		return m.menu.View()
	}
	leftWidth := (m.width * 30) / 100
	rightWidth := (m.width*70)/100 - 4
	leftBoxStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(m.height-2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderDefault).
		Background(bgSecondary).
		Foreground(textPrimary).
		Padding(1, 2)
	rightBoxStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(m.height-2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderDefault).
		Background(bgSecondary).
		Foreground(textPrimary).
		Padding(1, 2).
		Align(lipgloss.Center, lipgloss.Center)
	leftContent := leftBoxStyle.Render(m.form.View().Content)
	var rightContent string
	switch m.state {
	case stateMenu:
		return m.menu.View()
	case stateForm:
		initText := titleh1Style.Render("Seleccioná los parámetros a la izquierda") + "\n" + secondaryTextStyle().Render("y presioná ENTER para simular")
		rightContent = rightBoxStyle.Render(initText)
	case stateLoading:
		spinnerView := m.spinner.View() + " " + secondaryTextStyle().Render("Calculando simulación...")
		rightContent = rightBoxStyle.Render(spinnerView)
	case stateResults:
		theo, _ := ComputeTheoreticalStats(m.activeDistribution, m.distParams)
		probs, _ := ComputeProbabilities(m.activeDistribution, m.distParams)
		comparison := fmt.Sprintf(
			"%s\n\n"+
				"%-18s %10s %10s %10s\n"+
				"%-18s %10.4f %10.4f %10.4f\n"+
				"%-18s %10.4f %10.4f %10.4f\n"+
				"%-18s %10.4f %10.4f %10.4f\n"+
				"%-18s %10s %10d %10s\n",
			titleh1Style.Render("Comparación Simulado vs Teórico"),
			"Concepto", "Teórico", "Simulado", "Diferencia",
			"Media", theo.Avg, m.empiricalStats.Avg, theo.Avg-m.empiricalStats.Avg,
			"Varianza", theo.Variance, m.empiricalStats.Variance, theo.Variance-m.empiricalStats.Variance,
			"Desv. Estándar", theo.StdDev, math.Sqrt(m.empiricalStats.Variance), theo.StdDev-math.Sqrt(m.empiricalStats.Variance),
			"Tamaño muestra", "—", m.empiricalStats.Count, "—",
		)

		// Para continuas: f(x) es densidad, no probabilidad
		pxLabel := "P(X = x)"
		if !isDiscreteDistribution(m.activeDistribution) {
			pxLabel = "f(x)    "
		}
		probSection := fmt.Sprintf(
			"\n%s\n"+
				"%s  = %.6f\n"+
				"P(X ≤ x)  = %.6f\n"+
				"P(X > x)  = %.6f\n",
			titleh1Style.Render("Probabilidades para tu x"),
			pxLabel, probs.PX, probs.PLe, probs.PGt,
		)

		chartSection := ""
		if m.chartView != "" {
			chartSection = "\n" + titleh1Style.Render("Gráfica") + "\n" + m.chartView
		}

		exportSection := ""
		if m.exportMsg != "" {
			exportSection = "\n" + warningStyle.Render(m.exportMsg)
		}

		resultsView := comparison + probSection + chartSection + exportSection + "\n" + mutedStyle.Render("[ESC] volver  [e] PNG  [s] SVG  [c] CSV")

		resultsStyle := lipgloss.NewStyle().
			Width(rightWidth).
			Height(m.height-2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderDefault).
			Background(bgSecondary).
			Foreground(textPrimary).
			Padding(1, 2).
			Align(lipgloss.Left, lipgloss.Top)
		rightContent = resultsStyle.Render(resultsView)
	default:
		rightContent = rightBoxStyle.Render(errorTextStyle().Render("Estado desconocido"))
	}
	v := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent))
	return v
}

func NewMainModel() MainModel {
	model := MainModel{}

	model.styles = newStyles(false)
	model.menu = NewMenuModel()
	model.state = stateMenu
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle
	model.spinner = s
	return model
}
