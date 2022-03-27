package statusandprompt

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/utils"
)

// StatusAndPrompt is a resizing model which displays a submodel and a status bar.  When the start prompt
// event is received, focus will be torn away and the user will be given a prompt the enter text.
type StatusAndPrompt struct {
	model         layout.ResizingModel
	statusMessage string
	pendingInput  *startPromptMsg
	textInput     textinput.Model
	width         int
}

func New(model layout.ResizingModel, initialMsg string) StatusAndPrompt {
	textInput := textinput.New()
	return StatusAndPrompt{model: model, statusMessage: initialMsg, textInput: textInput}
}

func (s StatusAndPrompt) Init() tea.Cmd {
	return s.model.Init()
}

func (s StatusAndPrompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case setStatusMsg:
		s.statusMessage = string(msg)
	case startPromptMsg:
		if s.pendingInput != nil {
			// ignore, already in an input
			return s, nil
		}

		s.textInput.Prompt = msg.prompt
		s.textInput.Focus()
		s.textInput.SetValue("")
		s.pendingInput = &msg
		return s, nil
	case tea.KeyMsg:
		if s.pendingInput != nil {
			switch msg.String() {
			case "ctrl+c", "esc":
				s.pendingInput = nil
			case "enter":
				pendingInput := s.pendingInput
				s.pendingInput = nil

				return s, pendingInput.onDone(s.textInput.Value())
			}
		}
	}

	if s.pendingInput != nil {
		var cc utils.CmdCollector

		newTextInput, cmd := s.textInput.Update(msg)
		cc.Add(cmd)
		s.textInput = newTextInput

		if _, isKey := msg.(tea.Key); !isKey {
			s.model = cc.Collect(s.model.Update(msg)).(layout.ResizingModel)
		}

		return s, cc.Cmd()
	}

	newModel, cmd := s.model.Update(msg)
	s.model = newModel.(layout.ResizingModel)
	return s, cmd
}

func (s StatusAndPrompt) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, s.model.View(), s.viewStatus())
}

func (s StatusAndPrompt) Resize(w, h int) layout.ResizingModel {
	s.width = w
	submodelHeight := h - lipgloss.Height(s.viewStatus())
	s.model = s.model.Resize(w, submodelHeight)
	return s
}

func (s StatusAndPrompt) viewStatus() string {
	if s.pendingInput != nil {
		return s.textInput.View()
	}
	return s.statusMessage
}