package ui

import (
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Starbhein/DistCalc/internal/core/stats"
)

type sessionState int

const (
	stateMenu sessionState = iota
	stateForm
	stateLoading
	stateResults
)

type MainModel struct {
	styles         styles
	darkBG         bool
	state          sessionState
	menu           MenuModel
	form           FormModel
	width, height  int
	spinner        spinner.Model
	empiricalStats stats.EmpiricalStats
	chartBuffer    []float64
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrc+c", "q":
			return m, tea.Quit
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
		_, errM := Parser(msg.Parameters)
		if errM.error != nil {
			return m, func() tea.Msg {
				return errM
			}
		}
		m.state = stateLoading
		return m, func() tea.Msg {
			time.Sleep(2 * time.Second)
			return MsgSimulationSuccess{}
		}
	case MsgSimulationSuccess:
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
		BorderForeground(lipgloss.Color("#2A1F4A")).Padding(1, 2)
	rightBoxStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(m.height-2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#2A1F4A")).Padding(1, 2).
		Align(lipgloss.Center, lipgloss.Center)
	leftContent := leftBoxStyle.Render(m.form.View().Content)
	var rightContent string
	switch m.state {
	case stateMenu:
		return m.menu.View()
	case stateForm:
		initText := titleh1Style.Render("Selecciona los parámetros a la izquierda\n y presiona ENTER para simular")
		rightContent = rightBoxStyle.Render(initText)
	case stateLoading:
		spinnerView := m.spinner.View() + "Calculando simulación"
		rightContent = rightBoxStyle.Render(spinnerView)
	case stateResults:
		rightContent = rightBoxStyle.Render("Los resultados son los siguientes")
	default:
		rightContent = rightBoxStyle.Render("Estado desconocido")
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
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))
	model.spinner = s
	return model
}
