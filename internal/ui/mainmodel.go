package ui

import tea "charm.land/bubbletea/v2"

type sessionState int

const (
	stateMenu sessionState = iota
	stateForm
	stateLoading
	stateResults
)

type MainModel struct {
	styles        styles
	darkBG        bool
	state         sessionState
	menu          MenuModel
	form          FormModel
	width, height int
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
		return m, nil
	case MsgForm:
		res, errM := Parser(msg.Parameters)
		if errM.error != nil {
			return nil, func() tea.Msg {
				return errM
			}
		}
		return m, func() tea.Msg {}

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
	)
}

func (m MainModel) View() tea.View {
	switch m.state {
	case stateMenu:
		return m.menu.View()
		// case stateForm:
		// return m.form.View()
	default:
		return tea.NewView("Unknown state...")
	}
}

func NewMainModel() MainModel {
	model := MainModel{}

	model.styles = newStyles(false)
	model.menu = NewMenuModel()
	model.state = stateMenu

	return model
}
