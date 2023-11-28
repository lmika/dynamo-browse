package scriptmanager

import (
	"context"
	"fmt"
	"regexp"

	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"github.com/risor-io/risor/object"
)

var (
	validKeyBindingNames = regexp.MustCompile(`^[-a-zA-Z0-9_]+$`)
)

type extModule struct {
	scriptPlugin *ScriptPlugin
}

func (m *extModule) register() *object.Module {
	return object.NewBuiltinsModule("ext", map[string]object.Object{
		"command":       object.NewBuiltin("command", m.command),
		"key_binding":   object.NewBuiltin("key_binding", m.keyBinding),
		"related_items": object.NewBuiltin("related_items", m.relatedItem),
	})
}

func (m *extModule) command(ctx context.Context, args ...object.Object) object.Object {
	thisEnv := scriptEnvFromCtx(ctx)

	if err := require("ext.command", 2, args); err != nil {
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

		newEnv := thisEnv
		ctx = ctxWithScriptEnv(ctx, newEnv)

		res, err := callFn(ctx, fnRes, objArgs)
		if err != nil {
			return errors.Errorf("command error '%v':%v - %v", m.scriptPlugin.name, cmdName, err)
		} else if object.IsError(res) {
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
	thisEnv := scriptEnvFromCtx(ctx)

	if err := require("ext.key_binding", 3, args); err != nil {
		return err
	}

	bindingName, err := object.AsString(args[0])
	if err != nil {
		return err
	} else if !validKeyBindingNames.MatchString(bindingName) {
		return object.NewError(errors.New("value error: binding name must match regexp [-a-zA-Z0-9_]+"))
	}

	options, err := object.AsMap(args[1])
	if err != nil {
		return err
	}

	var defaultKey string
	if strVal, isStrVal := options.Get("default").(*object.String); isStrVal {
		defaultKey = strVal.Value()
	}

	fnRes, isFnRes := args[2].(*object.Function)
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

		newEnv := thisEnv
		ctx = ctxWithScriptEnv(ctx, newEnv)

		res, err := callFn(ctx, fnRes, objArgs)
		if err != nil {
			return errors.Errorf("command error '%v':%v - %v", m.scriptPlugin.name, bindingName, err)
		} else if object.IsError(res) {
			errObj := res.(*object.Error)
			return errors.Errorf("command error '%v':%v - %v", m.scriptPlugin.name, bindingName, errObj.Inspect())
		}
		return nil
	}

	fullBindingName := fmt.Sprintf("ext.%v.%v", m.scriptPlugin.name, bindingName)

	if m.scriptPlugin.definedKeyBindings == nil {
		m.scriptPlugin.definedKeyBindings = make(map[string]*Command)
		m.scriptPlugin.keyToKeyBinding = make(map[string]string)
	}

	m.scriptPlugin.definedKeyBindings[fullBindingName] = &Command{plugin: m.scriptPlugin, cmdFn: newCommand}
	m.scriptPlugin.keyToKeyBinding[defaultKey] = fullBindingName
	return nil
}

func (m *extModule) relatedItem(ctx context.Context, args ...object.Object) object.Object {
	thisEnv := scriptEnvFromCtx(ctx)

	var (
		tableName  string
		callbackFn *object.Function
	)
	if err := bindArgs("ext.related_items", args, &tableName, &callbackFn); err != nil {
		return err
	}

	callFn, hasCallFn := object.GetCallFunc(ctx)
	if !hasCallFn {
		return object.NewError(errors.New("no callFn found in context"))
	}

	// TEMP
	newHandler := func(ctx context.Context, rs *models.ResultSet, index int) ([]relatedItem, error) {
		newEnv := thisEnv
		newEnv.options = m.scriptPlugin.scriptService.options
		ctx = ctxWithScriptEnv(ctx, newEnv)

		res, err := callFn(ctx, callbackFn, []object.Object{
			newItemProxy(newResultSetProxy(rs), index),
		})

		if err != nil {
			return nil, errors.Errorf("script error '%v':related_item - %v", m.scriptPlugin.name, err)
		} else if object.IsError(res) {
			errObj := res.(*object.Error)
			return nil, errors.Errorf("script error '%v':related_item - %v", m.scriptPlugin.name, errObj.Inspect())
		}

		// TODO: map from list of maps -> slice of relItems

		return nil
	}
	// END TEMP

	newHandler()

	return nil
}
