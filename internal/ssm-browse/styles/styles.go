package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/statusandprompt"
)

type Styles struct {
	Frames          frame.Style
	StatusAndPrompt statusandprompt.Style
}

var DefaultStyles = Styles{
	Frames: frame.Style{
		ActiveTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#c144ff")),
		InactiveTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#d1d1d1")),
	},
	StatusAndPrompt: statusandprompt.Style{
		ModeLine: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#d1d1d1")),
	},
}
