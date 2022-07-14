package itemdisplay

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/utils"
)

type Model struct {
	baseMode tea.Model
}

func New(baseMode tea.Model) *Model {
	return &Model{
		baseMode: baseMode,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cc utils.CmdCollector

	m.baseMode = cc.Collect(m.baseMode.Update(msg))
	return m, cc.Cmd()
}

func (m *Model) View() string {
	return m.baseMode.View()
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.baseMode = layout.Resize(m.baseMode, w, h)
	return m
}
