package uimodels

import "context"

type uiContextKeyType struct{}

var uiContextKey = uiContextKeyType{}

func Ctx(ctx context.Context) UIContext {
	uiCtx, _ := ctx.Value(uiContextKey).(UIContext)
	return uiCtx
}

func WithContext(ctx context.Context, uiContext UIContext) context.Context {
	return context.WithValue(ctx, uiContextKey, uiContext)
}
