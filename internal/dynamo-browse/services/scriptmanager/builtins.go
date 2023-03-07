/**
 * Builtins adopted and modified from Taramin
 * Copyright (c) 2022 Curtis Myzie
 */

package scriptmanager

import (
	"context"
	"github.com/cloudcmds/tamarin/object"
	"log"
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
