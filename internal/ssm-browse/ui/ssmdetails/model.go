package ssmdetails

import (
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/ssm-browse/models"
	"strings"
)

type Model struct {
	frameTitle frame.FrameTitle
	viewport   viewport.Model
	w, h       int

	// model state
	hasSelectedItem bool
	selectedItem    *models.SSMParameter
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

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) SetSelectedItem(item *models.SSMParameter) {
	m.selectedItem = item

	if m.selectedItem != nil {
		var viewportContents strings.Builder
		fmt.Fprintf(&viewportContents, "Name: %v\n\n", item.Name)
		fmt.Fprintf(&viewportContents, "Type: TODO\n\n")
		fmt.Fprintf(&viewportContents, "%v\n", item.Value)

		m.viewport.SetContent(viewportContents.String())
	} else {
		m.viewport.SetContent("(no parameter selected)")
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
