package layout

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/utils"
	"strings"
)

// VBox is a model which will display its children vertically.
type VBox struct {
	children []ResizingModel
}

func NewVBox(children ...ResizingModel) VBox {
	return VBox{children: children}
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
	childrenHeight := h / len(vb.children)
	lastChildRem := h % len(vb.children)
	for i, c := range vb.children {
		if i == len(vb.children)-1 {
			vb.children[i] = c.Resize(w, childrenHeight+lastChildRem)
		} else {
			vb.children[i] = c.Resize(w, childrenHeight)
		}
	}
	return vb
}
