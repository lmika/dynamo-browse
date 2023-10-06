package scriptmanager

import (
	"context"
	"strings"

	"github.com/risor-io/risor/object"
)

type uiModule struct {
	uiService UIService
}

func (um *uiModule) print(ctx context.Context, args ...object.Object) object.Object {
	var msg strings.Builder
	for _, arg := range args {
		if arg == nil {
			continue
		}

		switch a := arg.(type) {
		case *object.String:
			msg.WriteString(a.Value())
		default:
			msg.WriteString(a.Inspect())
		}
	}

	um.uiService.PrintMessage(ctx, msg.String())
	return object.Nil
}

func (um *uiModule) prompt(ctx context.Context, args ...object.Object) object.Object {
	if err := require("ui.prompt", 1, args); err != nil {
		return err
	}

	msg, _ := object.AsString(args[0])
	respChan := um.uiService.Prompt(ctx, msg)

	select {
	case resp, hasResp := <-respChan:
		if hasResp {
			return object.NewString(resp)
		} else {
			return object.Nil
		}
	case <-ctx.Done():
		return object.NewError(ctx.Err())
	}
}

func (um *uiModule) register() *object.Module {
	return object.NewBuiltinsModule("ui", map[string]object.Object{
		"print":  object.NewBuiltin("print", um.print),
		"prompt": object.NewBuiltin("prompt", um.prompt),
	})
}
