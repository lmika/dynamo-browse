package commandctrl

import (
	"bufio"
	"bytes"
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"ucl.lmika.dev/ucl"

	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/shellwords"
)

const commandsCategory = "commands"

type CommandController struct {
	uclInst            *ucl.Inst
	historyProvider    IterProvider
	commandList        *CommandList
	msgSender          func(tea.Msg)
	lookupExtensions   []CommandLookupExtension
	completionProvider CommandCompletionProvider
}

func NewCommandController(historyProvider IterProvider) *CommandController {
	cc := &CommandController{
		historyProvider:  historyProvider,
		commandList:      nil,
		lookupExtensions: nil,
	}
	cc.uclInst = ucl.New(
		ucl.WithOut(ucl.LineHandler(cc.printLine)),
		ucl.WithMissingBuiltinHandler(cc.cmdInvoker),
	)
	return cc
}

func (c *CommandController) AddCommands(ctx *CommandList) {
	ctx.parent = c.commandList
	c.commandList = ctx
}

func (c *CommandController) SetMessageSender(msg func(tea.Msg)) {
	c.msgSender = msg
}

func (c *CommandController) AddCommandLookupExtension(ext CommandLookupExtension) {
	c.lookupExtensions = append(c.lookupExtensions, ext)
}

func (c *CommandController) SetCommandCompletionProvider(provider CommandCompletionProvider) {
	c.completionProvider = provider
}

func (c *CommandController) Prompt() tea.Msg {
	return events.PromptForInputMsg{
		Prompt:  ":",
		History: c.historyProvider.Iter(context.Background(), commandsCategory),
		OnDone: func(value string) tea.Msg {
			return c.Execute(value)
		},
		// TEMP
		OnTabComplete: func(value string) (string, bool) {
			if c.completionProvider == nil {
				return "", false
			}

			if strings.HasPrefix(value, "sa ") || strings.HasPrefix(value, "da ") {
				tokens := shellwords.Split(strings.TrimSpace(value))
				lastToken := tokens[len(tokens)-1]

				options := c.completionProvider.AttributesWithPrefix(lastToken)
				if len(options) == 1 {
					return value[:len(value)-len(lastToken)] + options[0], true
				}
			}
			return "", false
		},
		// END TEMP
	}
}

func (c *CommandController) Execute(commandInput string) tea.Msg {
	return c.execute(ExecContext{FromFile: false}, commandInput)
}

func (c *CommandController) execute(ctx ExecContext, commandInput string) tea.Msg {
	input := strings.TrimSpace(commandInput)
	if input == "" {
		return nil
	}

	res, err := c.uclInst.Eval(context.Background(), commandInput)
	if err != nil {
		return events.Error(err)
	}

	if teaMsg, ok := res.(teaMsgWrapper); ok {
		return teaMsg.msg
	}
	return nil
}

func (c *CommandController) Alias(commandName string) Command {
	return func(ctx ExecContext, args ucl.CallArgs) tea.Msg {
		command := c.lookupCommand(commandName)
		if command == nil {
			return events.Error(errors.New("no such command: " + commandName))
		}

		return command(ctx, args)
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

func (c *CommandController) cmdInvoker(ctx context.Context, name string, args ucl.CallArgs) (any, error) {
	command := c.lookupCommand(name)
	if command == nil {
		return nil, errors.New("no such command: " + name)
	}

	res := command(ExecContext{}, args)
	if errMsg, isErrMsg := res.(events.ErrorMsg); isErrMsg {
		return nil, errMsg
	}
	return teaMsgWrapper{res}, nil
}

func (c *CommandController) printLine(s string) {
	if c.msgSender != nil {
		c.msgSender(events.StatusMsg(s))
	}
}

type teaMsgWrapper struct {
	msg tea.Msg
}
