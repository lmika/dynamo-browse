/**
 * Builtins adopted and modified from Taramin
 * Copyright (c) 2022 Curtis Myzie
 */

package scriptmanager

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/risor-io/risor/object"
)

func printBuiltin(ctx context.Context, args ...object.Object) object.Object {
	env := scriptEnvFromCtx(ctx)
	prefix := "script " + env.filename + ":"

	values := make([]interface{}, len(args)+1)
	values[0] = prefix
	for i, arg := range args {
		switch arg := arg.(type) {
		case *object.String:
			values[i+1] = arg.Value()
		default:
			values[i+1] = arg.Inspect()
		}
	}
	log.Println(values...)
	return object.Nil
}

func printfBuiltin(ctx context.Context, args ...object.Object) object.Object {
	env := scriptEnvFromCtx(ctx)
	prefix := "script " + env.filename + ":"

	numArgs := len(args)
	if numArgs < 1 {
		return object.Errorf("type error: printf() takes 1 or more arguments (%d given)", len(args))
	}
	format, err := object.AsString(args[0])
	if err != nil {
		return err
	}
	var values = []interface{}{prefix}
	for _, arg := range args[1:] {
		switch arg := arg.(type) {
		case *object.String:
			values = append(values, arg.Value())
		default:
			values = append(values, arg.Interface())
		}
	}
	log.Printf("%s "+format, values...)
	return object.Nil
}

// This is taken from the args package
func require(funcName string, count int, args []object.Object) *object.Error {
	nArgs := len(args)
	if nArgs != count {
		if count == 1 {
			return object.Errorf(
				fmt.Sprintf("type error: %s() takes exactly 1 argument (%d given)",
					funcName, nArgs))
		}
		return object.Errorf(
			fmt.Sprintf("type error: %s() takes exactly %d arguments (%d given)",
				funcName, count, nArgs))
	}
	return nil
}

func bindArgs(funcName string, args []object.Object, bindArgs ...any) *object.Error {
	if err := require(funcName, len(bindArgs), args); err != nil {
		return err
	}

	for i, bindArg := range bindArgs {
		switch t := bindArg.(type) {
		case *string:
			str, err := object.AsString(args[i])
			if err != nil {
				return err
			}

			*t = str
		case **object.Function:
			fnRes, isFnRes := args[i].(*object.Function)
			if !isFnRes {
				return object.NewError(errors.Errorf("expected arg %v to be a function, was %T", i, bindArg))
			}

			*t = fnRes
		default:
			return object.NewError(errors.Errorf("unhandled arg type %v", i))
		}
	}
	return nil
}
