package main

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Starbhein/DistCalc/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.NewMainModel())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
