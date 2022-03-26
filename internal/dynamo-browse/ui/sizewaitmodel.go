package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"log"
)

// sizeWaitModel is a model which waits until the first screen size message comes through.  It then creates the
// submodel and delegates calls to that model
type sizeWaitModel struct {
	constr func(width, height int) tea.Model
	model  tea.Model
}

func newSizeWaitModel(constr func(width, height int) tea.Model) tea.Model {
	return sizeWaitModel{constr: constr}
}

func (s sizeWaitModel) Init() tea.Cmd {
	return nil
}

func (s sizeWaitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		log.Println("got window size message")
		if s.model == nil {
			log.Println("creating model")
			s.model = s.constr(m.Width, m.Height)
			s.model.Init()
		}
	}

	var submodelCmds tea.Cmd
	if s.model != nil {
		log.Println("starting update")
		s.model, submodelCmds = s.model.Update(msg)
		log.Println("ending update")
	}
	return s, submodelCmds
}

func (s sizeWaitModel) View() string {
	if s.model == nil {
		return ""
	}
	return s.model.View()
}
