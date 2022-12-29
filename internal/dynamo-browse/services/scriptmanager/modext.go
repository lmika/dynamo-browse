package scriptmanager

import (
	"context"
	"github.com/cloudcmds/tamarin/arg"
	"github.com/cloudcmds/tamarin/object"
	"github.com/cloudcmds/tamarin/scope"
	"github.com/pkg/errors"
)

type extModule struct {
	scriptPlugin *ScriptPlugin
}

func (m *extModule) register(scp *scope.Scope) {
	modScope := scope.New(scope.Opts{})
	mod := &object.Module{Name: "ext", Scope: modScope}

	modScope.AddBuiltins([]*object.Builtin{
		{Name: "command", Module: mod, Fn: m.command},
	})

	scp.Declare("ext", mod, true)
}

func (m *extModule) command(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("ext.command", 2, args); err != nil {
		return err
	}

	cmdName, err := object.AsString(args[0])
	if err != nil {
		return err
	}
	fnRes, isFnRes := args[1].(*object.Function)
	if !isFnRes {
		return object.NewError("expected second arg to be a function")
	}

	callFn, hasCallFn := object.GetCallFunc(ctx)
	if !hasCallFn {
		return object.NewError("no callFn found in context")
	}

	newCommand := func(ctx context.Context, args []string) error {
		objArgs := make([]object.Object, len(args))
		for i, a := range args {
			objArgs[i] = object.NewString(a)
		}

		// TODO: this should be on a separate thread
		res := callFn(ctx, fnRes.Scope, fnRes, objArgs)
		if object.IsError(res) {
			errObj := res.(*object.Error)
			return errors.Errorf("command error '%v':%v - %v", m.scriptPlugin.name, cmdName, errObj.Message)
		}
		return nil
	}

	if m.scriptPlugin.definedCommands == nil {
		m.scriptPlugin.definedCommands = make(map[string]Command)
	}
	m.scriptPlugin.definedCommands[cmdName] = newCommand
	return nil
}
