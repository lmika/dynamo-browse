package layout

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/utils"
	"strconv"
	"strings"
)

// ResizingModel is a model that handles resizing events.  The submodel will not get WindowSizeMessages but will
// guarantee to receive at least one resize event before the initial view.
type ResizingModel interface {
	tea.Model
	Resize(w, h int) ResizingModel
}

// Model takes a tea-model and displays it as a resizing model.  The model will be
// displayed with all the available space provided
func Model(m tea.Model) ResizingModel {
	return &teaModel{submodel: m}
}

type teaModel struct {
	submodel tea.Model
	w, h     int
}

func (t teaModel) Init() tea.Cmd {
	return t.submodel.Init()
}

func (t teaModel) Update(msg tea.Msg) (m tea.Model, cmd tea.Cmd) {
	t.submodel, cmd = t.submodel.Update(msg)
	return t, cmd
}

func (t teaModel) View() string {
	subview := t.submodel.View() + " (h: " + strconv.Itoa(t.h) + "\n"
	subviewHeight := lipgloss.Height(subview)
	subviewVPad := strings.Repeat("\n", utils.Max(t.h-subviewHeight-1, 0))
	return lipgloss.JoinVertical(lipgloss.Top, subview, subviewVPad)
}

func (t teaModel) Resize(w, h int) ResizingModel {
	t.w, t.h = w, h
	return t
}
