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
	mod := object.NewModule("ext", modScope)

	modScope.AddBuiltins([]*object.Builtin{
		object.NewBuiltin("command", m.command, mod),
		object.NewBuiltin("key_binding", m.keyBinding, mod),
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
		return object.NewError(errors.New("expected second arg to be a function"))
	}

	callFn, hasCallFn := object.GetCallFunc(ctx)
	if !hasCallFn {
		return object.NewError(errors.New("no callFn found in context"))
	}

	// This command function will be executed by the script scheduler
	newCommand := func(ctx context.Context, args []string) error {
		objArgs := make([]object.Object, len(args))
		for i, a := range args {
			objArgs[i] = object.NewString(a)
		}

		ctx = ctxWithOptions(ctx, m.scriptPlugin.scriptService.options)

		res := callFn(ctx, fnRes.Scope(), fnRes, objArgs)
		if object.IsError(res) {
			errObj := res.(*object.Error)
			return errors.Errorf("command error '%v':%v - %v", m.scriptPlugin.name, cmdName, errObj.Inspect())
		}
		return nil
	}

	if m.scriptPlugin.definedCommands == nil {
		m.scriptPlugin.definedCommands = make(map[string]*Command)
	}
	m.scriptPlugin.definedCommands[cmdName] = &Command{plugin: m.scriptPlugin, cmdFn: newCommand}
	return nil
}

func (m *extModule) keyBinding(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("ext.key_binding", 3, args); err != nil {
		return err
	}

	bindingName, err := object.AsString(args[0])
	if err != nil {
		return err
	}
	fnRes, isFnRes := args[1].(*object.Function)
	if !isFnRes {
		return object.NewError(errors.New("expected second arg to be a function"))
	}

	callFn, hasCallFn := object.GetCallFunc(ctx)
	if !hasCallFn {
		return object.NewError(errors.New("no callFn found in context"))
	}

	// This command function will be executed by the script scheduler
	newCommand := func(ctx context.Context, args []string) error {
		objArgs := make([]object.Object, len(args))
		for i, a := range args {
			objArgs[i] = object.NewString(a)
		}

		ctx = ctxWithOptions(ctx, m.scriptPlugin.scriptService.options)

		res := callFn(ctx, fnRes.Scope(), fnRes, objArgs)
		if object.IsError(res) {
			errObj := res.(*object.Error)
			return errors.Errorf("command error '%v':%v - %v", m.scriptPlugin.name, bindingName, errObj.Inspect())
		}
		return nil
	}

	if m.scriptPlugin.definedKeyBindings == nil {
		m.scriptPlugin.definedKeyBindings = make(map[string]*Command)
	}
	m.scriptPlugin.definedKeyBindings[bindingName] = &Command{plugin: m.scriptPlugin, cmdFn: newCommand}
	return nil
}
