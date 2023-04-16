package dynamoitemview

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/itemrenderer"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/styles"
	"strings"
)

type Model struct {
	ready               bool
	frameTitle          frame.FrameTitle
	viewport            viewport.Model
	itemRendererService *itemrenderer.Service
	w, h                int

	// model state
	currentResultSet *models.ResultSet
	selectedItem     models.Item
}

func New(itemRendererService *itemrenderer.Service, uiStyles styles.Styles) *Model {
	return &Model{
		itemRendererService: itemRendererService,
		frameTitle:          frame.NewFrameTitle("Item", false, uiStyles.Frames),
		viewport:            viewport.New(100, 100),
	}
}

func (*Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NewItemSelected:
		m.currentResultSet = msg.ResultSet
		m.selectedItem = msg.Item
		m.updateViewportToSelectedMessage()
		return m, nil
	}
	return m, nil
}

func (m *Model) View() string {
	if !m.ready {
		return ""
	}
	return lipgloss.JoinVertical(lipgloss.Top, m.frameTitle.View(), m.viewport.View())
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.w, m.h = w, h
	if !m.ready {
		m.viewport = viewport.New(w, h-m.frameTitle.HeaderHeight())
		m.viewport.SetContent("")
		m.ready = true
	} else {
		m.viewport.Width = w
		m.viewport.Height = h - m.frameTitle.HeaderHeight()
	}
	m.frameTitle.Resize(w, h)
	return m
}

func (m *Model) updateViewportToSelectedMessage() {
	if m.selectedItem == nil {
		m.viewport.SetContent("")
		return
	}

	viewportContent := &strings.Builder{}
	m.itemRendererService.RenderItem(viewportContent, m.selectedItem, m.currentResultSet, false)
	m.viewport.Width = m.w
	m.viewport.Height = m.h - m.frameTitle.HeaderHeight()
	m.viewport.SetContent(viewportContent.String())
}
