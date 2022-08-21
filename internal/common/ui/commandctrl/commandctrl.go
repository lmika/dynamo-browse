package commandctrl

import (
	"bufio"
	"bytes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"log"
	"strings"

	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/shellwords"
)

type CommandController struct {
	commandList    *CommandContext
	missingCommand MissingCommand
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

func (c *CommandController) SetMissingCommand(missingCommand MissingCommand) {
	c.missingCommand = missingCommand
}

func (c *CommandController) Prompt() tea.Msg {
	return events.PromptForInputMsg{
		Prompt: ":",
		OnDone: func(value string) tea.Msg {
			return c.Execute(value)
		},
	}
}

func (c *CommandController) Execute(commandInput string) tea.Msg {
	input := strings.TrimSpace(commandInput)
	if input == "" {
		return nil
	}

	tokens := shellwords.Split(input)
	command := c.lookupCommand(tokens[0])
	if command != nil {
		return command(tokens[1:])
	}

	if c.missingCommand != nil {
		command = c.missingCommand(tokens[0])
		if command != nil {
			return command(tokens[1:])
		}
	}

	log.Println("No such command: ", tokens)
	return events.Error(errors.New("no such command: " + tokens[0]))
}

func (c *CommandController) Alias(commandName string) Command {
	return func(args []string) tea.Msg {
		command := c.lookupCommand(commandName)
		if command == nil {
			log.Println("No such command: ", commandName)
			return events.Error(errors.New("no such command: " + commandName))
		}

		return command(args)
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

func (c *CommandController) ExecuteFile(file []byte, filename string) error {
	scnr := bufio.NewScanner(bytes.NewReader(file))
	for scnr.Scan() {
		line := strings.TrimSpace(scnr.Text())
		if line == "" {
			continue
		} else if line[0] == '#' {
			continue
		}

		c.Execute(line) // TODO: deal with errors
	}
	return scnr.Err()
}
