package dialogprompt

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
)

type Model struct {
	compositor *layout.Compositor
}

func New(model layout.ResizingModel) *Model {
	m := &Model{
		compositor: layout.NewCompositor(model),
	}
	// TEMP
	m.compositor.SetOverlay(&dialogModel{}, 5, 5, 30, 12)
	return m
}

func (m *Model) Init() tea.Cmd {
	return m.compositor.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := m.compositor.Update(msg)
	m.compositor = newModel.(*layout.Compositor)
	return m, cmd
}

func (m *Model) View() string {
	return m.compositor.View()
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.compositor = m.compositor.Resize(w, h).(*layout.Compositor)
	return m
}
