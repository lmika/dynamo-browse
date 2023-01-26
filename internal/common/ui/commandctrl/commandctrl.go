package commandctrl

import (
	"bufio"
	"bytes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/shellwords"
)

type CommandController struct {
	commandList      *CommandList
	lookupExtensions []CommandLookupExtension
}

func NewCommandController() *CommandController {
	return &CommandController{
		commandList:      nil,
		lookupExtensions: nil,
	}
}

func (c *CommandController) AddCommands(ctx *CommandList) {
	ctx.parent = c.commandList
	c.commandList = ctx
}

func (c *CommandController) AddCommandLookupExtension(ext CommandLookupExtension) {
	c.lookupExtensions = append(c.lookupExtensions, ext)
}

func (c *CommandController) Prompt() tea.Msg {
	return events.PromptForInputMsg{
		Prompt:  ":",
		History: tempCommandHistory,
		OnDone: func(value string) tea.Msg {
			return c.Execute(value)
		},
	}
}

// TEMP
var tempCommandHistory = new(commandHistory)

type commandHistory struct {
	history []string
}

func (ch *commandHistory) Item(idx int) string {
	return ch.history[idx]
}

func (ch *commandHistory) PutItem(onDoneMsg tea.Msg, item string) {
	if _, isErrMsg := onDoneMsg.(events.ErrorMsg); !isErrMsg {
		ch.history = append(ch.history, item)
	}
}

func (ch *commandHistory) Len() int {
	return len(ch.history)
}

// END TEMP

func (c *CommandController) Execute(commandInput string) tea.Msg {
	return c.execute(ExecContext{FromFile: false}, commandInput)
}

func (c *CommandController) execute(ctx ExecContext, commandInput string) tea.Msg {
	input := strings.TrimSpace(commandInput)
	if input == "" {
		return nil
	}

	tokens := shellwords.Split(input)
	command := c.lookupCommand(tokens[0])
	if command == nil {
		return events.Error(errors.New("no such command: " + tokens[0]))
	}

	return command(ctx, tokens[1:])
}

func (c *CommandController) Alias(commandName string, aliasArgs []string) Command {
	return func(ctx ExecContext, args []string) tea.Msg {
		command := c.lookupCommand(commandName)
		if command == nil {
			return events.Error(errors.New("no such command: " + commandName))
		}

		var allArgs []string
		if len(aliasArgs) > 0 {
			allArgs = append(append([]string{}, aliasArgs...), args...)
		} else {
			allArgs = args
		}
		return command(ctx, allArgs)
	}
}

func (c *CommandController) lookupCommand(name string) Command {
	for ctx := c.commandList; ctx != nil; ctx = ctx.parent {
		if cmd, ok := ctx.Commands[name]; ok {
			return cmd
		}
	}

	for _, exts := range c.lookupExtensions {
		if cmd := exts.LookupCommand(name); cmd != nil {
			return cmd
		}
	}
	return nil
}

func (c *CommandController) ExecuteFile(filename string) error {
	baseFilename := filepath.Base(filename)

	if rcFile, err := os.ReadFile(filename); err == nil {
		if err := c.executeFile(rcFile, baseFilename); err != nil {
			return errors.Wrapf(err, "error executing %v", filename)
		}
	} else {
		return errors.Wrapf(err, "error loading %v", filename)
	}
	return nil
}

func (c *CommandController) executeFile(file []byte, filename string) error {
	scnr := bufio.NewScanner(bytes.NewReader(file))

	lineNo := 0
	for scnr.Scan() {
		lineNo++
		line := strings.TrimSpace(scnr.Text())
		if line == "" {
			continue
		} else if line[0] == '#' {
			continue
		}

		msg := c.execute(ExecContext{FromFile: true}, line)
		switch m := msg.(type) {
		case events.ErrorMsg:
			log.Printf("%v:%v: error - %v", filename, lineNo, m.Error())
		case events.StatusMsg:
			log.Printf("%v:%v: %v", filename, lineNo, string(m))
		}
	}
	return scnr.Err()
}
