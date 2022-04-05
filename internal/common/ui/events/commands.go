package events

import (
	tea "github.com/charmbracelet/bubbletea"
	"log"
)

func Error(err error) tea.Msg {
	log.Println(err)
	return ErrorMsg(err)
}

func SetError(err error) tea.Cmd {
	return func() tea.Msg {
		return Error(err)
	}
}

func SetStatus(msg string) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg(msg)
	}
}

func PromptForInput(prompt string, onDone func(value string) tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		return PromptForInputMsg{
			Prompt: prompt,
			OnDone: onDone,
		}
	}
}

type MessageWithStatus interface {
	StatusMessage() string
}
