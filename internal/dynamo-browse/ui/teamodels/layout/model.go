package layout

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/utils"
	"strconv"
	"strings"
)

// ResizingModel is a model that handles resizing events.  The submodel will not get WindowSizeMessages but will
// guarantee to receive at least one resize event before the initial view.
type ResizingModel interface {
	tea.Model
	Resize(w, h int) ResizingModel
}

// Resize sends a resize message to the passed in model.  If m implements ResizingModel, then Resize is called;
// otherwise, m is returned without any messages.
func Resize(m tea.Model, w, h int) tea.Model {
	if rm, isRm := m.(ResizingModel); isRm {
		return rm.Resize(w, h)
	}
	return m
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

type ResizableModelHandler struct {
	new    func(w, h int) tea.Model
	resize func(m tea.Model, w, h int) tea.Model
	model  tea.Model
}

// NewResizableModelHandler takes a tea model that requires a with and height during construction
// and has a resize method, and wraps it as a resizing model.
func NewResizableModelHandler(newModel func(w, h int) tea.Model) ResizableModelHandler {
	return ResizableModelHandler{
		new: newModel,
	}
}

func (rmh ResizableModelHandler) WithResize(resizeFn func(m tea.Model, w, h int) tea.Model) ResizableModelHandler {
	rmh.resize = resizeFn
	return rmh
}

func (rmh ResizableModelHandler) Resize(w, h int) ResizingModel {
	if rmh.model == nil {
		rmh.model = rmh.new(w, h)
		// TODO: handle init
	} else if rmh.resize != nil {
		rmh.model = rmh.resize(rmh.model, w, h)
	}
	return rmh
}

func (rmh ResizableModelHandler) Init() tea.Cmd {
	return nil
}

func (rmh ResizableModelHandler) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if rmh.model == nil {
		return rmh, nil
	}

	newModel, cmd := rmh.model.Update(msg)
	rmh.model = newModel
	return rmh, cmd
}

func (rmh ResizableModelHandler) View() string {
	if rmh.model == nil {
		return ""
	}

	return rmh.model.View()
}
