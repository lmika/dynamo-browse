package tableselect

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/utils"
)

type Model struct {
	frameTitle       frame.FrameTitle
	listController   listController
	submodel         tea.Model
	pendingSelection *controllers.PromptForTableMsg
	isLoading        bool
	w, h             int
}

func New(submodel tea.Model) Model {
	frameTitle := frame.NewFrameTitle("Select table", false)
	return Model{frameTitle: frameTitle, submodel: submodel}
}

func (m Model) Init() tea.Cmd {
	return m.submodel.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cc utils.CmdCollector
	switch msg := msg.(type) {
	case controllers.PromptForTableMsg:
		m.isLoading = false
		m.pendingSelection = &msg
		m.listController = newListController(msg.Tables, m.w, m.h-m.frameTitle.HeaderHeight())
		return m, nil
	case indicateLoadingTablesMsg:
		m.isLoading = true
		return m, nil
	case tea.KeyMsg:
		if m.pendingSelection != nil {
			switch msg.String() {
			case "enter":
				if m.listController.list.FilterState() != list.Filtering {
					var sel controllers.PromptForTableMsg
					sel, m.pendingSelection = *m.pendingSelection, nil

					return m, sel.OnSelected(m.listController.list.SelectedItem().(tableItem).name)
				}
			}

			m.listController = cc.Collect(m.listController.Update(msg)).(listController)
			return m, cc.Cmd()
		}
	}

	m.submodel = cc.Collect(m.submodel.Update(msg))
	return m, cc.Cmd()
}

func (m Model) View() string {
	if m.pendingSelection != nil {
		return lipgloss.JoinVertical(lipgloss.Top, m.frameTitle.View(), m.listController.View())
	} else if m.isLoading {
		return lipgloss.JoinVertical(lipgloss.Top, m.frameTitle.View(), "Loading tables")
	}

	return m.submodel.View()
}

func (m Model) shouldShow() bool {
	return m.pendingSelection != nil || m.isLoading
}

func (m Model) Resize(w, h int) layout.ResizingModel {
	m.w, m.h = w, h
	m.submodel = layout.Resize(m.submodel, w, h)

	m.frameTitle.Resize(w, h)
	if m.pendingSelection != nil {
		m.listController = m.listController.Resize(w, h-m.frameTitle.HeaderHeight()).(listController)
	}
	return m
}
