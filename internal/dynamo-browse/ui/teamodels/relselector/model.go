package relselector

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/controllers"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/utils"
)

const (
	overlayWidth  = 50
	overlayHeight = 30
)

type Model struct {
	subModel   tea.Model
	compositor *layout.Compositor
	listModel  *listModel
	w, h       int
}

func New(subModel tea.Model) *Model {
	compositor := layout.NewCompositor(subModel)
	listModel := newListModel()

	return &Model{
		subModel:   subModel,
		listModel:  listModel,
		compositor: compositor,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.compositor.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cc utils.CmdCollector
	switch msg := msg.(type) {
	case controllers.ShowRelatedItemsOverlay:
		m.compositor.SetOverlay(m.listModel, m.w/2-overlayWidth/2, m.h/2-overlayHeight/2, overlayWidth, overlayHeight)
	case tea.KeyMsg:
		m.compositor = cc.Collect(m.compositor.Update(msg)).(*layout.Compositor)
	default:
		m.subModel = cc.Collect(m.subModel.Update(msg)).(tea.Model)
	}
	return m, cc.Cmd()
}

func (m *Model) View() string {
	return m.compositor.View()
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.w, m.h = w, h
	m.compositor.MoveOverlay(m.w/2-overlayWidth/2, m.h/2-overlayHeight/2)
	m.listModel.Resize(w, h)
	m.subModel = layout.Resize(m.subModel, w, h)
	m.listModel = layout.Resize(m.listModel, w, h).(*listModel)
	return m
}

func (m *Model) SelectorVisible() bool {
	return m.compositor.HasOverlay()
}
