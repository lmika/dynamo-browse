package commandctrl

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"log"
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
		log.Println("No such command: ", tokens)
		return events.SetError(errors.New("no such command: " + tokens[0]))
	}

	return command(tokens[1:])
}

func (c *CommandController) Alias(commandName string) Command {
	return func(args []string) tea.Cmd {
		command := c.lookupCommand(commandName)
		if command == nil {
			log.Println("No such command: ", commandName)
			return events.SetError(errors.New("no such command: " + commandName))
		}

		if len(args) > 1 {
			return command(args[1:])
		}
		return command([]string{})
	}
}

func (c *CommandController) lookupCommand(name string) Command {
	for ctx := c.commandList; ctx != nil; ctx = ctx.parent {
		log.Printf("Looking in command list: %v", c.commandList)
		if cmd, ok := ctx.Commands[name]; ok {
			return cmd
		}
	}
	return nil
}
