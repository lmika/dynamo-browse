package events

import (
	tea "github.com/charmbracelet/bubbletea"
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
	Prompt   string
	History  HistoryProvider
	OnDone   func(value string) tea.Msg
	OnCancel func() tea.Msg
}

type HistoryProvider interface {
	// Len returns the number of historical items
	Len() int

	// Item returns the historical item at index 'idx', where items are chronologically ordered such that the
	// item at 0 is the oldest item.
	Item(idx int) string

	// PutItem adds an item to the history
	PutItem(onDoneMsg tea.Msg, item string)
}
