package scriptmanager

import (
	"context"
	"github.com/cloudcmds/tamarin/arg"
	"github.com/cloudcmds/tamarin/object"
	"github.com/cloudcmds/tamarin/scope"
	"strings"
)

type uiModule struct {
	uiService UIService
}

func (um *uiModule) print(ctx context.Context, args ...object.Object) object.Object {
	var msg strings.Builder
	for _, arg := range args {
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
	if err := arg.Require("ui.prompt", 1, args); err != nil {
		return err
	}

	msg, _ := object.AsString(args[0])
	respChan := um.uiService.Prompt(ctx, msg)

	select {
	case resp, hasResp := <-respChan:
		if hasResp {
			return object.NewString(resp)
		} else {
			return object.NewError(ctx.Err())
		}
	case <-ctx.Done():
		return object.NewError(ctx.Err())
	}
}

func (um *uiModule) register(scp *scope.Scope) {
	modScope := scope.New(scope.Opts{})
	mod := object.NewModule("ui", modScope)

	modScope.AddBuiltins([]*object.Builtin{
		object.NewBuiltin("print", um.print, mod),
		object.NewBuiltin("prompt", um.prompt, mod),
	})

	scp.Declare("ui", mod, true)
}
