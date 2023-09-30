package events

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services"
)

// Error indicates that an error occurred
type ErrorMsg error

// Message indicates that a message should be shown to the user
type StatusMsg string

type WrappedStatusMsg struct {
	Message StatusMsg
	Next    tea.Msg
}

// ModeMessage indicates that the mode should be changed to the following
type ModeMessage string

// PromptForInput indicates that the context is requesting a line of input
type PromptForInputMsg struct {
	Prompt        string
	History       services.HistoryProvider
	OnDone        func(value string) tea.Msg
	OnCancel      func() tea.Msg
	OnTabComplete func(value string) (string, bool)
}
