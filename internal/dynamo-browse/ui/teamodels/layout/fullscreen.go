package layout

import tea "github.com/charmbracelet/bubbletea"

// FullScreen returns a model which will allocate the resizing model the entire height and width of the screen.
func FullScreen(rm ResizingModel) tea.Model {
	return fullScreenModel{submodel: rm}
}

type fullScreenModel struct {
	w, h     int
	submodel ResizingModel
	ready    bool
}

func (f fullScreenModel) Init() tea.Cmd {
	return f.submodel.Init()
}

func (f fullScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.ready = true
		f.submodel = f.submodel.Resize(msg.Width, msg.Height)
		return f, nil
	}

	newSubModel, cmd := f.submodel.Update(msg)
	f.submodel = newSubModel.(ResizingModel)
	return f, cmd
}

func (f fullScreenModel) View() string {
	if !f.ready {
		return ""
	}
	return f.submodel.View()
}
