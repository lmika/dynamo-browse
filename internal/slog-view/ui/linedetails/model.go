package linedetails

import (
	"encoding/json"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/slog-view/models"
)

type Model struct {
	frameTitle frame.FrameTitle
	viewport   viewport.Model
	w, h       int

	// model state
	focused bool
	selectedItem    *models.LogLine
}

func New() *Model {
	viewport := viewport.New(0, 0)
	viewport.SetContent("")
	return &Model{
		frameTitle: frame.NewFrameTitle("Item", false),
		viewport:   viewport,
	}
}

func (*Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetFocused(newFocused bool) {
	m.focused = newFocused
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.focused {
			newModel, cmd := m.viewport.Update(msg)
			m.viewport = newModel
			return m, cmd
		}
	}

	return m, nil
}

func (m *Model) SetSelectedItem(item *models.LogLine) {
	m.selectedItem = item

	if m.selectedItem != nil {
		if formattedJson, err := json.MarshalIndent(item.JSON, "", "   "); err == nil {
			m.viewport.SetContent(string(formattedJson))
		} else {
			m.viewport.SetContent("(not json)")
		}
	} else {
		m.viewport.SetContent("(no line)")
	}
}

func (m *Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, m.frameTitle.View(), m.viewport.View())
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.w, m.h = w, h
	m.frameTitle.Resize(w, h)
	m.viewport.Width = w
	m.viewport.Height = h - m.frameTitle.HeaderHeight()
	return m
}
