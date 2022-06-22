package dynamoitemview

import (
	"fmt"
	"github.com/lmika/awstools/internal/dynamo-browse/models/itemrender"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/styles"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
)

var (
	activeHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#4479ff"))

	fieldTypeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#2B800C", Dark: "#73C653"})
	metaInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
)

type Model struct {
	ready      bool
	frameTitle frame.FrameTitle
	viewport   viewport.Model
	w, h       int

	// model state
	currentResultSet *models.ResultSet
	selectedItem     models.Item
}

func New(uiStyles styles.Styles) *Model {
	return &Model{
		frameTitle: frame.NewFrameTitle("Item", false, uiStyles.Frames),
		viewport:   viewport.New(100, 100),
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
	}

	viewportContent := &strings.Builder{}
	tabWriter := tabwriter.NewWriter(viewportContent, 0, 1, 1, ' ', 0)
	for _, colName := range m.currentResultSet.Columns {
		if r := m.selectedItem.Renderer(colName); r != nil {
			m.renderItem(tabWriter, "", colName, r)
		}
	}

	tabWriter.Flush()
	m.viewport.Width = m.w
	m.viewport.Height = m.h - m.frameTitle.HeaderHeight()
	m.viewport.SetContent(viewportContent.String())
}

func (m *Model) renderItem(w io.Writer, prefix string, name string, r itemrender.Renderer) {
	fmt.Fprintf(w, "%s%v\t%s\t%s%s\n",
		prefix, name, fieldTypeStyle.Render(r.TypeName()), r.StringValue(), metaInfoStyle.Render(r.MetaInfo()))
	if subitems := r.SubItems(); len(subitems) > 0 {
		for _, si := range subitems {
			m.renderItem(w, prefix+"  ", si.Key, si.Value)
		}
	}
}
