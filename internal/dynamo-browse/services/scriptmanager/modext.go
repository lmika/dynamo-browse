package scriptmanager

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/queryexpr"
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

		itr, objErr := object.AsIterator(res)
		if err != nil {
			return nil, objErr.Value()
		}

		var relItems []relatedItem
		for next, hasNext := itr.Next(); hasNext; next, hasNext = itr.Next() {
			var newRelItem relatedItem

			itemMap, objErr := object.AsMap(next)
			if err != nil {
				return nil, objErr.Value()
			}

			labelName, objErr := object.AsString(itemMap.Get("label"))
			if objErr != nil {
				continue
			}
			newRelItem.label = labelName

			var tableStr = ""
			if itemMap.Get("table") != object.Nil {
				tableStr, objErr = object.AsString(itemMap.Get("table"))
				if objErr != nil {
					continue
				}
			}
			newRelItem.table = tableStr

			if selectFn, ok := itemMap.Get("on_select").(*object.Function); ok {
				newRelItem.onSelect = func() error {
					thisNewEnv := thisEnv
					thisNewEnv.options = m.scriptPlugin.scriptService.options
					ctx = ctxWithScriptEnv(ctx, thisNewEnv)

					res, err := callFn(ctx, selectFn, []object.Object{})
					if err != nil {
						return errors.Errorf("rel error '%v' - %v", m.scriptPlugin.name, err)
					} else if object.IsError(res) {
						errObj := res.(*object.Error)
						return errors.Errorf("rel error '%v' - %v", m.scriptPlugin.name, errObj.Inspect())
					}
					return nil
				}
			} else {
				queryExprStr, objErr := object.AsString(itemMap.Get("query"))
				if objErr != nil {
					continue
				}

				query, err := queryexpr.Parse(queryExprStr)
				if err != nil {
					continue
				}

				// Placeholders
				if argsVal, isArgsValMap := object.AsMap(itemMap.Get("args")); isArgsValMap == nil {
					namePlaceholders := make(map[string]string)
					valuePlaceholders := make(map[string]types.AttributeValue)

					for k, val := range argsVal.Value() {
						switch v := val.(type) {
						case *object.String:
							namePlaceholders[k] = v.Value()
							valuePlaceholders[k] = &types.AttributeValueMemberS{Value: v.Value()}
						case *object.Int:
							valuePlaceholders[k] = &types.AttributeValueMemberN{Value: fmt.Sprint(v.Value())}
						case *object.Float:
							valuePlaceholders[k] = &types.AttributeValueMemberN{Value: fmt.Sprint(v.Value())}
						case *object.Bool:
							valuePlaceholders[k] = &types.AttributeValueMemberBOOL{Value: v.Value()}
						case *object.NilType:
							valuePlaceholders[k] = &types.AttributeValueMemberNULL{Value: true}
						default:
							continue
						}
					}

					query = query.WithNameParams(namePlaceholders).WithValueParams(valuePlaceholders)
				}
				newRelItem.query = query
			}

			relItems = append(relItems, newRelItem)
		}

		return relItems, nil
	}

	m.scriptPlugin.relatedItems = append(m.scriptPlugin.relatedItems, &relatedItemBuilder{
		table:          tableName,
		itemProduction: newHandler,
	})

	return nil
}
