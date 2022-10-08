package statusandprompt

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/audax/internal/common/sliceutils"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/utils"
	"log"
)

// StatusAndPrompt is a resizing model which displays a submodel and a status bar.  When the start prompt
// event is received, focus will be torn away and the user will be given a prompt the enter text.
type StatusAndPrompt struct {
	model          layout.ResizingModel
	style          Style
	modeLine       string
	statusMessage  string
	spinner        spinner.Model
	spinnerVisible bool
	pendingInput   *events.PromptForInputMsg
	textInput      textinput.Model
	width          int
}

type Style struct {
	ModeLine lipgloss.Style
}

func New(model layout.ResizingModel, initialMsg string, style Style) *StatusAndPrompt {
	textInput := textinput.New()
	return &StatusAndPrompt{
		model:         model,
		style:         style,
		statusMessage: initialMsg,
		modeLine:      "",
		spinner:       spinner.New(spinner.WithSpinner(spinner.Line)),
		textInput:     textInput,
	}
}

func (s *StatusAndPrompt) Init() tea.Cmd {
	return tea.Batch(
		s.model.Init(),
		//s.spinner.Tick,
	)
}

func (s *StatusAndPrompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cc utils.CmdCollector

	switch msg := msg.(type) {
	case events.ErrorMsg:
		s.statusMessage = "Error: " + msg.Error()
	case events.StatusMsg:
		s.statusMessage = string(msg)
	case events.WrappedStatusMsg:
		s.statusMessage = string(msg.Message)
		cc.Add(func() tea.Msg { return msg.Next })
	case events.ForegroundJobUpdate:
		if msg.JobRunning {
			s.spinnerVisible = true
			s.statusMessage = msg.JobStatus
			cc.Add(s.spinner.Tick)
		} else {
			s.spinnerVisible = false
		}
	case events.ModeMessage:
		s.modeLine = string(msg)
	case events.MessageWithStatus:
		if hasModeMessage, ok := msg.(events.MessageWithMode); ok {
			s.modeLine = hasModeMessage.ModeMessage()
		}
		s.statusMessage = msg.StatusMessage()
	case events.PromptForInputMsg:
		if s.pendingInput != nil {
			// ignore, already in an input
			return s, nil
		}

		s.textInput.Prompt = msg.Prompt
		s.textInput.Focus()
		s.textInput.SetValue("")
		s.pendingInput = &msg
		return s, nil
	case tea.KeyMsg:
		if s.pendingInput != nil {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				s.pendingInput = nil
			case tea.KeyEnter:
				pendingInput := s.pendingInput
				s.pendingInput = nil

				return s, func() tea.Msg {
					m := pendingInput.OnDone(s.textInput.Value())
					log.Printf("return msg type = %T", m)
					return m
				}
			default:
				if msg.Type == tea.KeyRunes {
					msg.Runes = sliceutils.Filter(msg.Runes, func(r rune) bool { return r != '\x0d' && r != '\x0a' })
				}
				newTextInput, cmd := s.textInput.Update(msg)
				s.textInput = newTextInput
				return s, cmd
			}
		} else {
			s.statusMessage = ""
		}
	}

	if s.spinnerVisible {
		s.spinner = cc.Collect(s.spinner.Update(msg)).(spinner.Model)
	}
	s.model = cc.Collect(s.model.Update(msg)).(layout.ResizingModel)
	return s, cc.Cmd()
}

func (s *StatusAndPrompt) InPrompt() bool {
	return s.pendingInput != nil
}

func (s *StatusAndPrompt) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, s.model.View(), s.viewStatus())
}

func (s *StatusAndPrompt) Resize(w, h int) layout.ResizingModel {
	s.width = w
	submodelHeight := h - lipgloss.Height(s.viewStatus())
	s.model = s.model.Resize(w, submodelHeight)
	return s
}

func (s *StatusAndPrompt) viewStatus() string {
	modeLine := s.style.ModeLine.Render(lipgloss.PlaceHorizontal(s.width, lipgloss.Left, s.modeLine, lipgloss.WithWhitespaceChars(" ")))

	var statusLine string
	if s.pendingInput != nil {
		statusLine = s.textInput.View()
	} else {
		statusLine = s.statusMessage
	}

	if s.spinnerVisible {
		statusLine = lipgloss.JoinHorizontal(lipgloss.Left, s.spinner.View(), " ", statusLine)
	}

	return lipgloss.JoinVertical(lipgloss.Top, modeLine, statusLine)
}
