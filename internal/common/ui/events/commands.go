package events

import tea "github.com/charmbracelet/bubbletea"

func Error(err error) tea.Msg {
	return ErrorMsg(err)
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
