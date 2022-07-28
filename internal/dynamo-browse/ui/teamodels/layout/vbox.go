package layout

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/utils"
)

// VBox is a model which will display its children vertically.
type VBox struct {
	boxSize  BoxSize
	children []ResizingModel
}

func NewVBox(boxSize BoxSize, children ...ResizingModel) VBox {
	return VBox{boxSize: boxSize, children: children}
}

func (vb VBox) Init() tea.Cmd {
	var cc utils.CmdCollector
	for _, c := range vb.children {
		cc.Collect(c, c.Init())
	}
	return cc.Cmd()
}

func (vb VBox) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cc utils.CmdCollector
	for i, c := range vb.children {
		vb.children[i] = cc.Collect(c.Update(msg)).(ResizingModel)
	}
	return vb, cc.Cmd()
}

func (vb VBox) View() string {
	sb := new(strings.Builder)
	for i, c := range vb.children {
		if i > 0 {
			sb.WriteRune('\n')
		}
		sb.WriteString(c.View())
	}
	return sb.String()
}

func (vb VBox) Resize(w, h int) ResizingModel {
	for i, c := range vb.children {
		childHeight := vb.boxSize.childSize(i, len(vb.children), h)
		vb.children[i] = c.Resize(w, childHeight)
	}
	return vb
}
