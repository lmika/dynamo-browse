package colselector

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/lmika/audax/internal/dynamo-browse/ui/keybindings"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/utils"
)

type Model struct {
	keyBinding        *keybindings.TableKeyBinding
	columnsController *controllers.ColumnsController
	subModel          tea.Model
	colListModel      *colListModel
	compositor        *layout.Compositor
}

func New(submodel tea.Model, keyBinding *keybindings.TableKeyBinding, columnsController *controllers.ColumnsController) *Model {
	colListModel := newColListModel(keyBinding, columnsController)

	compositor := layout.NewCompositor(submodel)

	return &Model{
		keyBinding:        keyBinding,
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
		m.compositor.SetOverlay(m.colListModel, 6, 4, 50, 20)
	case controllers.HideColumnOverlay:
		m.compositor.ClearOverlay()
	case controllers.ColumnsUpdated:
		m.colListModel.refreshTable()
		m.subModel = cc.Collect(m.subModel.Update(msg))
	case tea.KeyMsg:
		m.compositor = cc.Collect(m.compositor.Update(msg)).(*layout.Compositor)
	default:
		m.subModel = cc.Collect(m.subModel.Update(msg))
	}
	return m, cc.Cmd()
}

func (m *Model) View() string {
	return m.compositor.View()
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.subModel = layout.Resize(m.subModel, w, h)
	m.colListModel = layout.Resize(m.colListModel, w, h).(*colListModel)
	return m
}

func (m *Model) ColSelectorVisible() bool {
	return m.compositor.HasOverlay()
}
