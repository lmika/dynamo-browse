package commandctrl

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"

	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/shellwords"
)

type CommandController struct {
	commandList *CommandContext
}

func NewCommandController() *CommandController {
	return &CommandController{
		commandList: nil,
	}
}

func (c *CommandController) AddCommands(ctx *CommandContext) {
	ctx.parent = c.commandList
	c.commandList = ctx
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
	command := c.lookupCommand(tokens[0])
	if command == nil {
		return events.SetStatus("no such command: " + tokens[0])
	}

	return command(tokens[1:])
}

func (c *CommandController) lookupCommand(name string) Command {
	for ctx := c.commandList; ctx != nil; ctx = ctx.parent {
		if cmd, ok := ctx.Commands[name]; ok {
			return cmd
		}
	}
	return nil
}