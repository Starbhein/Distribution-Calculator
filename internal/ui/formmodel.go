package ui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/Starbhein/DistCalc/internal/core/distributions/registry"
)

type FormModel struct {
	inputs             []textinput.Model
	focusIndex         int
	activeDistribution string
	styles             styles
	errorMap           map[int]string
	generalError       string
	isCLTMode          bool
}
type MsgForm struct {
	Parameters []string
}

func (form FormModel) IsComplete() bool {
	for _, v := range form.inputs {
		if v.Value() == "" {
			return false
		}
	}
	return true
}

func (form FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "tab":
			if form.focusIndex >= len(form.inputs)-1 {
				form.inputs[form.focusIndex].Blur()
				form.focusIndex = 0
				form.inputs[form.focusIndex].Focus()
				return form, nil
			}
			form.inputs[form.focusIndex].Blur()
			form.focusIndex++
			form.inputs[form.focusIndex].Focus()
			return form, nil
		case "up", "shift+tab":
			if form.focusIndex <= 0 {
				form.inputs[form.focusIndex].Blur()
				form.focusIndex = len(form.inputs) - 1
				form.inputs[form.focusIndex].Focus()
				return form, nil
			}

			form.inputs[form.focusIndex].Blur()
			form.focusIndex--
			form.inputs[form.focusIndex].Focus()
			return form, nil
		case "enter":

			if form.IsComplete() {
				return form, func() tea.Msg {
					values := make([]string, len(form.inputs))
					for i, v := range form.inputs {
						values[i] = v.Value()
					}
					return MsgForm{
						Parameters: values,
					}
				}
			}

		default:
			delete(form.errorMap, form.focusIndex)
			form.generalError = ""
		}
	case errorMessage:
		if msg.index >= 0 {
			if form.errorMap == nil {
				form.errorMap = make(map[int]string)
			}
			form.errorMap[msg.index] = msg.error.Error()
		} else {
			form.generalError = msg.error.Error()
		}
	}
	var cmd tea.Cmd
	form.inputs[form.focusIndex], cmd = form.inputs[form.focusIndex].Update(msg)
	return form, cmd
}

func (form FormModel) View() tea.View {
	var model strings.Builder
	for i := range form.inputs {
		model.WriteString(form.inputs[i].View())

		if errM, ok := form.errorMap[i]; ok {
			errorLabel := errorLabelStyle.Render(errM)

			model.WriteString(errorLabel)
		}
		model.WriteString("\n\n")
	}
	boton := "[ ENTER para iniciar simulación ]"
	var btn string
	if form.IsComplete() {
		btn = buttonActiveStyle.Render(boton)
	} else {
		btn = buttonDisabledStyle.Render(boton)
	}
	model.WriteString(btn)
	model.WriteString("\n")
	if form.generalError != "" {
		model.WriteString(errorLabelStyle.Render("Error: " + form.generalError))
		model.WriteString("\n")
	}
	model.WriteString(mutedStyle.Render("[ESC] volver al menú"))
	return tea.NewView(model.String())
}

func (form FormModel) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
	)
}

func (form *FormModel) BuildInputs(distribution string) {
	form.inputs = []textinput.Model{}
	form.activeDistribution = distribution
	createInput := func(placeholder string) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		s := ti.Styles()
		s.Focused.Placeholder = focusedStyle
		s.Focused.Text = focusedStyle
		s.Focused.Prompt = focusedStyle
		s.Cursor.Shape = tea.CursorBar
		s.Cursor.Blink = true
		s.Blurred.Prompt = blurredStyle
		s.Blurred.Placeholder = mutedStyle
		ti.CharLimit = 15
		ti.SetWidth(20)
		ti.SetStyles(s)
		return ti
	}
	// Form labels come from the registry spec (design §1.3 — ParamLabels
	// replaces this switch); the trailing "X (x)" input is simulation-only.
	if spec, ok := registry.ByName(distribution); ok {
		for _, label := range spec.ParamLabels {
			form.inputs = append(form.inputs, createInput(label))
		}
		if !form.isCLTMode {
			form.inputs = append(form.inputs, createInput("X (x)"))
		}
	}
	if !form.isCLTMode {
		inputDefault := createInput("Tamaño de la muestra a simular")
		inputDefault.SetValue("1000")
		form.inputs = append(form.inputs, inputDefault)
	}
	form.focusIndex = 0
	form.inputs[0].Focus()
}
