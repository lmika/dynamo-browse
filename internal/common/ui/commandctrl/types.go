package commandctrl

import tea "github.com/charmbracelet/bubbletea"

type Command func(args []string) tea.Cmd

func NoArgCommand(cmd tea.Cmd) Command {
	return func(args []string) tea.Cmd {
		return cmd
	}
}
