package statusandprompt

import tea "github.com/charmbracelet/bubbletea"

type setStatusMsg string

type startPromptMsg struct {
	prompt string
	onDone func(val string) tea.Cmd
}

func SetStatus(newStatus string) tea.Cmd {
	return func() tea.Msg {
		return setStatusMsg(newStatus)
	}
}

func Prompt(prompt string, onDone func(val string) tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		return startPromptMsg{
			prompt: prompt,
			onDone: onDone,
		}
	}
}
