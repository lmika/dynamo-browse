package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/statusandprompt"
)

type Styles struct {
	ItemView        ItemViewStyle
	Frames          frame.Style
	StatusAndPrompt statusandprompt.Style
}

type ItemViewStyle struct {
	FieldType lipgloss.Style
	MetaInfo  lipgloss.Style
}

var DefaultStyles = Styles{
	ItemView: ItemViewStyle{
		FieldType: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#2B800C", Dark: "#73C653"}),
		MetaInfo:  lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
	},
	Frames: frame.Style{
		ActiveTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#4479ff")),
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
