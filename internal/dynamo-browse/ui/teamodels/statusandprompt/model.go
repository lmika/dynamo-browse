package statusandprompt

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/dynamo-browse/internal/common/sliceutils"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/utils"
)

// StatusAndPrompt is a resizing model which displays a submodel and a status bar.  When the start prompt
// event is received, focus will be torn away and the user will be given a prompt the enter text.
type StatusAndPrompt struct {
	model              layout.ResizingModel
	style              Style
	modeLine           string
	rightModeLine      string
	statusMessage      string
	spinner            spinner.Model
	spinnerVisible     bool
	pendingInput       *pendingInputState
	textInput          textinput.Model
	width, height      int
	lastModeLineHeight int
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
		rightModeLine: "",
		spinner:       spinner.New(spinner.WithSpinner(spinner.Line)),
		textInput:     textInput,
	}
}

func (s *StatusAndPrompt) Init() tea.Cmd {
	return tea.Batch(
		s.model.Init(),
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
		if rightModeMessage, ok := msg.(events.MessageWithRightMode); ok {
			s.rightModeLine = rightModeMessage.RightModeMessage()
		} else {
			s.rightModeLine = ""
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
		s.pendingInput = newPendingInputState(msg)
	case tea.KeyMsg:
		if s.pendingInput != nil {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				if s.pendingInput.originalMsg.OnCancel != nil {
					pendingInput := s.pendingInput
					cc.Add(func() tea.Msg {
						m := pendingInput.originalMsg.OnCancel()
						return m
					})
				}
				s.pendingInput = nil
			case tea.KeyTab:
				if tabCompletion := s.pendingInput.originalMsg.OnTabComplete; tabCompletion != nil {
					if completion, ok := tabCompletion(s.textInput.Value()); ok {
						s.textInput.SetValue(completion)
						s.textInput.SetCursor(len(s.textInput.Value()))
					}
				}
			case tea.KeyEnter:
				pendingInput := s.pendingInput
				s.pendingInput = nil

				m := pendingInput.originalMsg.OnDone(s.textInput.Value())

				return s, tea.Batch(
					events.SetTeaMessage(m),
					func() tea.Msg {
						if historyProvider := pendingInput.originalMsg.History; historyProvider != nil {
							if _, isErrMsg := m.(events.ErrorMsg); !isErrMsg {
								historyProvider.PutItem(s.textInput.Value())
							}
						}
						return nil
					},
				)
			case tea.KeyUp:
				if historyProvider := s.pendingInput.originalMsg.History; historyProvider != nil && historyProvider.Len() > 0 {
					if s.pendingInput.historyIdx < 0 {
						s.pendingInput.historyIdx = historyProvider.Len() - 1
					} else if s.pendingInput.historyIdx > 0 {
						s.pendingInput.historyIdx -= 1
					} else {
						s.pendingInput.historyIdx = 0
					}
					s.textInput.SetValue(historyProvider.Item(s.pendingInput.historyIdx))
					s.textInput.SetCursor(len(s.textInput.Value()))
				}
			case tea.KeyDown:
				if historyProvider := s.pendingInput.originalMsg.History; historyProvider != nil && historyProvider.Len() > 0 {
					if s.pendingInput.historyIdx >= 0 && s.pendingInput.historyIdx < historyProvider.Len()-1 {
						s.pendingInput.historyIdx += 1
					}
					s.textInput.SetValue(historyProvider.Item(s.pendingInput.historyIdx))
					s.textInput.SetCursor(len(s.textInput.Value()))
				}
			default:
				if msg.Type == tea.KeyRunes {
					msg.Runes = sliceutils.Filter(msg.Runes, func(r rune) bool { return r != '\x0d' && r != '\x0a' })
				}
			}

			s.textInput = cc.Collect(s.textInput.Update(msg)).(textinput.Model)
			return s, cc.Cmd()
		} else {
			s.statusMessage = ""
		}
	}

	if s.spinnerVisible {
		s.spinner = cc.Collect(s.spinner.Update(msg)).(spinner.Model)
	}
	s.model = cc.Collect(s.model.Update(msg)).(layout.ResizingModel)

	// If the height of the modeline has changed, request a relayout
	if s.lastModeLineHeight != lipgloss.Height(s.viewStatus()) {
		cc.Add(events.SetTeaMessage(layout.RequestLayout{}))
	}
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
	s.height = h
	s.lastModeLineHeight = lipgloss.Height(s.viewStatus())
	submodelHeight := h - s.lastModeLineHeight
	s.model = s.model.Resize(w, submodelHeight)
	return s
}

func (s *StatusAndPrompt) viewStatus() string {
	rightModeLine := s.style.ModeLine.Render(s.rightModeLine)
	modeLine := s.style.ModeLine.Render(
		lipgloss.PlaceHorizontal(s.width-lipgloss.Width(rightModeLine), lipgloss.Left, s.modeLine, lipgloss.WithWhitespaceChars(" ")),
	) + rightModeLine

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
