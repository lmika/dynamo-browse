package colselector

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/controllers"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/keybindings"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/utils"
)

const (
	overlayWidth  = 50
	overlayHeight = 25
)

type Model struct {
	columnsController *controllers.ColumnsController
	subModel          tea.Model
	colListModel      *colListModel
	compositor        *layout.Compositor
	w, h              int
}

func New(submodel tea.Model, keyBinding *keybindings.KeyBindings, columnsController *controllers.ColumnsController) *Model {
	colListModel := newColListModel(keyBinding, columnsController)

	compositor := layout.NewCompositor(submodel)

	return &Model{
		columnsController: columnsController,
		subModel:          submodel,
		compositor:        compositor,
		colListModel:      colListModel,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.subModel.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cc utils.CmdCollector
	switch msg := msg.(type) {
	case controllers.ShowColumnOverlay:
		m.colListModel.setColumnsFromModel(m.columnsController.Columns())
		m.compositor.SetOverlay(m.colListModel, m.w/2-overlayWidth/2, m.h/2-overlayHeight/2, overlayWidth, overlayHeight)
	case controllers.HideColumnOverlay:
		m.compositor.ClearOverlay()
	case controllers.ColumnsUpdated:
		m.colListModel.refreshTable()
		m.subModel = cc.Collect(m.subModel.Update(msg)).(tea.Model)
	case controllers.SetSelectedColumnInColSelector:
		m.compositor = cc.Collect(m.compositor.Update(msg)).(*layout.Compositor)
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
	m.subModel = layout.Resize(m.subModel, w, h)
	m.colListModel = layout.Resize(m.colListModel, w, h).(*colListModel)
	return m
}

func (m *Model) ColSelectorVisible() bool {
	return m.compositor.HasOverlay()
}
