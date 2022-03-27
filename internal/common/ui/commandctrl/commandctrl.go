package commandctrl

import (
	"context"
	"strings"

	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/common/ui/uimodels"
	"github.com/lmika/shellwords"
	"github.com/pkg/errors"
)

type CommandController struct {
	commands map[string]uimodels.Operation
}

func NewCommandController(commands map[string]uimodels.Operation) *CommandController {
	return &CommandController{
		commands: commands,
	}
}

func (c *CommandController) Prompt() uimodels.Operation {
	return uimodels.OperationFn(func(ctx context.Context) error {
		uiCtx := uimodels.Ctx(ctx)
		uiCtx.Send(events.PromptForInputMsg{
			Prompt: ":",
			// OnDone: c.Execute(),
		})
		return nil
	})
}

func (c *CommandController) Execute() uimodels.Operation {
	return uimodels.OperationFn(func(ctx context.Context) error {
		input := strings.TrimSpace(uimodels.PromptValue(ctx))
		if input == "" {
			return nil
		}

		tokens := shellwords.Split(input)
		command, ok := c.commands[tokens[0]]
		if !ok {
			return errors.New("no such command: " + tokens[0])
		}

		return command.Execute(WithCommandArgs(ctx, tokens[1:]))
	})
}
