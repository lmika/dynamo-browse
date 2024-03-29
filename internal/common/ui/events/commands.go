package events

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services"
	"log"
)

func Error(err error) tea.Msg {
	log.Println(err)
	return ErrorMsg(err)
}

func SetStatus(msg string) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg(msg)
	}
}

func SetTeaMessage(event tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return event
	}
}

func PromptForInput(prompt string, history services.HistoryProvider, onDone func(value string) tea.Msg) tea.Msg {
	return PromptForInputMsg{
		Prompt:  prompt,
		History: history,
		OnDone:  onDone,
	}
}

func Confirm(prompt string, onResult func(yes bool) tea.Msg) tea.Msg {
	return PromptForInput(prompt, nil, func(value string) tea.Msg {
		return onResult(value == "y")
	})
}

func ConfirmYes(prompt string, onYes func() tea.Msg) tea.Msg {
	return PromptForInput(prompt, nil, func(value string) tea.Msg {
		if value == "y" {
			return onYes()
		}
		return nil
	})
}

type MessageWithStatus interface {
	StatusMessage() string
}

type MessageWithMode interface {
	MessageWithStatus
	ModeMessage() string
}

type MessageWithRightMode interface {
	MessageWithStatus
	RightModeMessage() string
}
