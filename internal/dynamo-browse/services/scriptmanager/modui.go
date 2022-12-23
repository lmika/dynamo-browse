package scriptmanager

import (
	"context"
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
		if s, err := object.AsString(arg); err == nil {
			msg.WriteString(s)
		}
	}

	um.uiService.PrintMessage(msg.String())
	return object.Nil
}

func (um *uiModule) register(scp *scope.Scope) {
	modScope := scope.New(scope.Opts{})
	mod := &object.Module{Name: "ui", Scope: modScope}

	modScope.AddBuiltins([]*object.Builtin{
		{Name: "print", Module: mod, Fn: um.print},
	})

	scp.Declare("ui", mod, true)
}
