package commandctrl

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"

	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/shellwords"
)

type CommandController struct {
	commands map[string]Command
}

func NewCommandController(commands map[string]Command) *CommandController {
	return &CommandController{
		commands: commands,
	}
}

func (c *CommandController) Prompt() tea.Cmd {
	return func() tea.Msg {
		return events.PromptForInputMsg{
			Prompt: ":",
			OnDone: func(value string) tea.Cmd {
				return c.Execute(value)
			},
		}
	}
}

func (c *CommandController) Execute(commandInput string) tea.Cmd {
	input := strings.TrimSpace(commandInput)
	if input == "" {
		return nil
	}

	tokens := shellwords.Split(input)
	command, ok := c.commands[tokens[0]]
	if !ok {
		return events.SetStatus("no such command: " + tokens[0])
	}

	return command(tokens)
}
