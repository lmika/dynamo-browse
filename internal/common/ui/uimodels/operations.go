package uimodels

import "context"

type Operation interface {
	Execute(ctx context.Context) error
}

type OperationFn func(ctx context.Context) error

func (f OperationFn) Execute(ctx context.Context) error {
	return f(ctx)
}
