package ui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const distributionQuantity = 5

type styles struct {
	app           lipgloss.Style
	title         lipgloss.Style
	statusMessage lipgloss.Style
}

func newStyles(darkBG bool) styles {
	lightDark := lipgloss.LightDark(darkBG)

	return styles{
		app: lipgloss.NewStyle().
			Padding(1, 2).Align(lipgloss.Center),
		title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF2079")).
			Background(lipgloss.Color("#440BD4")).
			Padding(0, 1),
		statusMessage: lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#04005E"), lipgloss.Color("#04B575"))),
	}
}

type distributionOption struct {
	title       string
	description string
}
type MsgSelectedDistribution struct {
	Distribution string
}

func (d distributionOption) Title() string       { return d.title }
func (d distributionOption) Description() string { return d.description }
func (d distributionOption) FilterValue() string { return d.title }

type MenuModel struct {
	menu   list.Model
	styles styles
	darkBG bool
}

func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			option, ok := m.menu.SelectedItem().(distributionOption)
			if ok {
				return m, func() tea.Msg {
					return MsgSelectedDistribution{Distribution: option.title}
				}
			}
		}
	}
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)
	return m, cmd
}

func (m MenuModel) View() tea.View {
	v := tea.NewView(m.styles.app.Render(m.menu.View()))
	v.AltScreen = true
	return v
}

func NewMenuModel() MenuModel {
	model := MenuModel{}
	model.styles = newStyles(false)
	// distributionOptions := make([]list.Item, distributionQuantity)
	// for i := range distributionOptions {
	// 	distributionOptions[i] = distributionOption{title: "hola", description: "papus" + string(i)}
	// }
	distributionOptions := initDistributionOptions()
	model.menu = list.New(distributionOptions, list.NewDefaultDelegate(), 500, 500)
	model.menu.Title = "Calculadora de distribuciones de probabilidad"
	return model
}

func initDistributionOptions() []list.Item {
	options := []string{"Binomial", "Poisson", "Hypergeométrica", "Normal", "Exponencial"}
	descriptions := []string{
		"B~(x,n,p) where n*p <=5 ", "P~(x,λ) and P~(x,β)", "H~(x,m,k,N)", "N~(x),Z normalization", "e~(λ)",
	}
	res := make([]list.Item, distributionQuantity)
	for i := range res {
		res[i] = distributionOption{title: options[i], description: descriptions[i]}
	}
	return res
}

func (m MenuModel) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
	)
}
