package events

import (
	tea "github.com/charmbracelet/bubbletea"
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

func PromptForInput(prompt string, onDone func(value string) tea.Msg) tea.Msg {
	return PromptForInputMsg{
		Prompt: prompt,
		OnDone: onDone,
	}
}

func Confirm(prompt string, onYes func() tea.Msg) tea.Msg {
	return PromptForInput(prompt, func(value string) tea.Msg {
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
