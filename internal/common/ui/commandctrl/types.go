package commandctrl

import tea "github.com/charmbracelet/bubbletea"

type Command func(ctx ExecContext, args []string) tea.Msg

type MissingCommand func(name string) Command

func NoArgCommand(cmd tea.Cmd) Command {
	return func(ctx ExecContext, args []string) tea.Msg {
		return cmd()
	}
}

type CommandList struct {
	Commands map[string]Command

	parent *CommandList
}
