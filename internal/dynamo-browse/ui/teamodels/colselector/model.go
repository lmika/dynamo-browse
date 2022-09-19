package colselector

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/utils"
)

type Model struct {
	subModel     tea.Model
	colListModel *colListModel
	compositor   *layout.Compositor
}

func New(submodel tea.Model) *Model {
	colListModel := &colListModel{}

	compositor := layout.NewCompositor(submodel)
	compositor.SetOverlay(colListModel, 5, 5, 30, 10)

	return &Model{
		subModel:     submodel,
		compositor:   compositor,
		colListModel: colListModel,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.subModel.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cc utils.CmdCollector
	m.subModel = cc.Collect(m.subModel.Update(msg))
	return m, cc.Cmd()
}

func (m *Model) View() string {
	return m.compositor.View()
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.subModel = layout.Resize(m.subModel, w, h)
	return m
}
