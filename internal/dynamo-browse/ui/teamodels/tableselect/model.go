package tableselect

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/modal"
)

type Model struct {
	pendingSelection *showTableSelectMsg
	modal            modal.Modal
	w, h             int
}

func New(submodel tea.Model) Model {
	return Model{modal: modal.New(submodel)}
}

func (m Model) Init() tea.Cmd {
	return m.modal.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case showTableSelectMsg:
		m.pendingSelection = &msg
		m.modal.Push(newListController(m.w, m.h))
		return m, nil
	case tea.KeyMsg:
		if m.modal.Len() > 0 {
			switch msg.String() {
			case "enter":
				listController := m.modal.Pop().(listController)

				var sel showTableSelectMsg
				sel, m.pendingSelection = *m.pendingSelection, nil

				return m, sel.onSelected(listController.list.SelectedItem().(tableItem).name)
			}
		}
	}

	newModal, cmd := m.modal.Update(msg)
	m.modal = newModal.(modal.Modal)
	return m, cmd
}

func (m Model) View() string {
	return m.modal.View()
}

func (m Model) Resize(w, h int) layout.ResizingModel {
	m.w, m.h = w, h
	m.modal = layout.Resize(m.modal, w, h).(modal.Modal)
	return m
}
