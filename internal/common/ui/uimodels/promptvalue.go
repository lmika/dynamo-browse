package uimodels

import "context"

type promptValueKeyType struct {}
var promptValueKey = promptValueKeyType{}

func PromptValue(ctx context.Context) string {
	value, _ := ctx.Value(promptValueKey).(string)
	return value
}

func WithPromptValue(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, promptValueKey, value)
}
