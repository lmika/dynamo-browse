package layout

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/utils"
)

type ZStack struct {
	visibleModel tea.Model
	focusedModel tea.Model
	otherModels  []tea.Model
}

func NewZStack(visibleModel tea.Model, focusedModel tea.Model, otherModels ...tea.Model) ZStack {
	return ZStack{
		visibleModel: visibleModel,
		focusedModel: focusedModel,
		otherModels:  otherModels,
	}
}

func (vb ZStack) Init() tea.Cmd {
	var cc utils.CmdCollector
	cc.Collect(vb.visibleModel, vb.visibleModel.Init())
	cc.Collect(vb.focusedModel, vb.focusedModel.Init())
	for _, c := range vb.otherModels {
		cc.Collect(c, c.Init())
	}
	return cc.Cmd()
}

func (vb ZStack) Update(msg tea.Msg) (m tea.Model, cmd tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		// Only the focused model gets keyboard events
		vb.focusedModel, cmd = vb.focusedModel.Update(msg)
		return vb, cmd
	}

	// All other messages go to each model
	var cc utils.CmdCollector
	vb.visibleModel = cc.Collect(vb.visibleModel.Update(msg)).(tea.Model)
	vb.focusedModel = cc.Collect(vb.focusedModel.Update(msg)).(tea.Model)
	for i, c := range vb.otherModels {
		vb.otherModels[i] = cc.Collect(c.Update(msg)).(tea.Model)
	}
	return vb, cc.Cmd()
}

func (vb ZStack) View() string {
	return vb.visibleModel.View()
}

func (vb ZStack) Resize(w, h int) ResizingModel {
	vb.visibleModel = Resize(vb.visibleModel, w, h)
	vb.focusedModel = Resize(vb.focusedModel, w, h)
	for i := range vb.otherModels {
		vb.otherModels[i] = Resize(vb.otherModels[i], w, h)
	}
	return vb
}
