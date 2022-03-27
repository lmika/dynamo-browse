package frame

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/utils"
	"strings"
)

var (
	activeHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#4479ff"))

	inactiveHeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#d1d1d1"))
)

// Frame is a frame that appears in the
type Frame struct {
	header string
	active bool
	model  layout.ResizingModel
	width  int
}

func NewFrame(header string, active bool, model layout.ResizingModel) Frame {
	return Frame{header, active, model, 0}
}

func (f Frame) Init() tea.Cmd {
	return f.model.Init()
}

func (f Frame) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		// If frame is not active, do not receive key messages
		if !f.active {
			return f, nil
		}
	}

	newModel, cmd := f.model.Update(msg)
	f.model = newModel.(layout.ResizingModel)
	return f, cmd
}

func (f Frame) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, f.headerView(), f.model.View())
}

func (f Frame) Resize(w, h int) layout.ResizingModel {
	f.width = w
	headerHeight := lipgloss.Height(f.headerView())
	f.model = f.model.Resize(w, h-headerHeight)
	return f
}

func (f Frame) headerView() string {
	style := inactiveHeaderStyle
	if f.active {
		style = activeHeaderStyle
	}

	titleText := f.header
	title := style.Render(titleText)
	line := style.Render(strings.Repeat(" ", utils.Max(0, f.width-lipgloss.Width(title))))
	return lipgloss.JoinHorizontal(lipgloss.Left, title, line)
}
