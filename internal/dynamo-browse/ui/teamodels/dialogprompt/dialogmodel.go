package dialogprompt

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
)

var style = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63"))

type dialogModel struct {
	w, h        int
	borderStyle lipgloss.Style
}

func (d *dialogModel) Init() tea.Cmd {
	return nil
}

func (d *dialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return d, nil
}

func (d *dialogModel) View() string {
	return style.Width(d.w).Height(d.h).Render("Hello this is a test of some content")
}

func (d *dialogModel) Resize(w, h int) layout.ResizingModel {
	d.w, d.h = w-2, h-2
	return d
}
