package commandctrl

import "context"

type commandArgContextKeyType struct{}

var commandArgContextKey = commandArgContextKeyType{}

func WithCommandArgs(ctx context.Context, args []string) context.Context {
	return context.WithValue(ctx, commandArgContextKey, args)
}

func CommandArgs(ctx context.Context) []string {
	args, _ := ctx.Value(commandArgContextKey).([]string)
	return args
}
