package events

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Error indicates that an error occurred
type ErrorMsg error

// Message indicates that a message should be shown to the user
type StatusMsg string

// ModeMessage indicates that the mode should be changed to the following
type ModeMessage string

// PromptForInput indicates that the context is requesting a line of input
type PromptForInputMsg struct {
	Prompt string
	OnDone func(value string) tea.Msg
}
