package colselector

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/layout"
)

var style = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63"))

type colListModel struct {
}

func (c *colListModel) Init() tea.Cmd {
	return nil
}

func (c *colListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return c, nil
}

func (c *colListModel) View() string {
	return style.Width(28).Height(8).Render("Hello")
}

func (c *colListModel) Resize(w, h int) layout.ResizingModel {
	return c
}
