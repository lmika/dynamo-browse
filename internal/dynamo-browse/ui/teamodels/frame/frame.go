package frame

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/utils"
)

var (
	inactiveHeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#d1d1d1"))
)

// Frame is a frame that appears in the
type FrameTitle struct {
	header string
	active bool
	style  Style
	width  int
}

type Style struct {
	ActiveTitle   lipgloss.Style
	InactiveTitle lipgloss.Style
}

func NewFrameTitle(header string, active bool, style Style) FrameTitle {
	return FrameTitle{header, active, style, 0}
}

func (f *FrameTitle) SetTitle(title string) {
	f.header = title
}

func (f FrameTitle) View() string {
	return f.headerView()
}

func (f *FrameTitle) Resize(w, h int) {
	f.width = w
}

func (f FrameTitle) HeaderHeight() int {
	return lipgloss.Height(f.headerView())
}

func (f FrameTitle) headerView() string {
	style := f.style.InactiveTitle
	if f.active {
		style = f.style.ActiveTitle
	}

	titleText := f.header
	title := style.Render(titleText)
	line := style.Render(strings.Repeat(" ", utils.Max(0, f.width-lipgloss.Width(title))))
	return lipgloss.JoinHorizontal(lipgloss.Left, title, line)
}
