package commandctrl

import tea "github.com/charmbracelet/bubbletea"

type Command func(args []string) tea.Cmd

type MissingCommand func(name string) Command

func NoArgCommand(cmd tea.Cmd) Command {
	return func(args []string) tea.Cmd {
		return cmd
	}
}

type CommandContext struct {
	Commands map[string]Command

	parent *CommandContext
}
