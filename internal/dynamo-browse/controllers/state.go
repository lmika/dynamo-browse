package controllers

import (
	"context"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
)

type State struct {
	ResultSet    *models.ResultSet
	SelectedItem models.Item

	// InReadWriteMode indicates whether modifications can be made to the table
	InReadWriteMode bool
}

type stateContextKeyType struct{}

var stateContextKey = stateContextKeyType{}

func CurrentState(ctx context.Context) State {
	state, _ := ctx.Value(stateContextKey).(State)
	return state
}

func ContextWithState(ctx context.Context, state State) context.Context {
	return context.WithValue(ctx, stateContextKey, state)
}
