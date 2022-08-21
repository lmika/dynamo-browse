package commandctrl

import tea "github.com/charmbracelet/bubbletea"

type Command func(args []string) tea.Msg

type MissingCommand func(name string) Command

func NoArgCommand(cmd tea.Cmd) Command {
	return func(args []string) tea.Msg {
		return cmd()
	}
}

type CommandContext struct {
	Commands map[string]Command

	parent *CommandContext
}
