package fullviewlinedetails

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/slog-view/models"
	"github.com/lmika/awstools/internal/slog-view/ui/linedetails"
)

type Model struct {
	submodel    tea.Model
	lineDetails *linedetails.Model

	visible bool
}

func NewModel(submodel tea.Model, style frame.Style) *Model {
	return &Model{
		submodel:    submodel,
		lineDetails: linedetails.New(style),
	}
}

func (*Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.visible = false
			return m, nil
		}

		if m.visible {
			newModel, cmd := m.lineDetails.Update(msg)
			m.lineDetails = newModel.(*linedetails.Model)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.submodel, cmd = m.submodel.Update(msg)
	return m, cmd
}

func (m *Model) ViewItem(item *models.LogLine) {
	m.visible = true
	m.lineDetails.SetSelectedItem(item)
	m.lineDetails.SetFocused(true)
}

func (m *Model) View() string {
	if m.visible {
		return m.lineDetails.View()
	}
	return m.submodel.View()
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.submodel = layout.Resize(m.submodel, w, h)
	m.lineDetails = layout.Resize(m.lineDetails, w, h).(*linedetails.Model)
	return m
}
