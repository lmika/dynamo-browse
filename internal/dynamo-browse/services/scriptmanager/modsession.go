package scriptmanager

import (
	"context"
	"github.com/cloudcmds/tamarin/arg"
	"github.com/cloudcmds/tamarin/object"
	"github.com/cloudcmds/tamarin/scope"
	"github.com/pkg/errors"
)

type sessionModule struct {
	sessionService SessionService
}

func (um *sessionModule) query(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("session.query", 1, args); err != nil {
		return err
	}

	expr, objErr := object.AsString(args[0])
	if objErr != nil {
		return objErr
	}
	resp, err := um.sessionService.Query(ctx, expr)

	if err != nil {
		return object.NewErrResult(object.NewError(err))
	}
	return object.NewOkResult(&resultSetProxy{resultSet: resp})
}

func (um *sessionModule) resultSet(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("session.result_set", 0, args); err != nil {
		return err
	}

	rs := um.sessionService.ResultSet(ctx)
	if rs == nil {
		return object.Nil
	}
	return &resultSetProxy{resultSet: rs}
}

func (um *sessionModule) selectedItem(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("session.result_set", 0, args); err != nil {
		return err
	}

	rs := um.sessionService.ResultSet(ctx)
	idx := um.sessionService.SelectedItemIndex(ctx)
	if rs == nil || idx < 0 {
		return object.Nil
	}

	rsProxy := &resultSetProxy{resultSet: rs}
	return newItemProxy(rsProxy, idx)
}

func (um *sessionModule) setResultSet(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("session.set_result_set", 1, args); err != nil {
		return err
	}

	resultSetProxy, isResultSetProxy := args[0].(*resultSetProxy)
	if !isResultSetProxy {
		return object.NewError(errors.Errorf("type error: expected a resultsset (got %v)", args[0]))
	}

	um.sessionService.SetResultSet(ctx, resultSetProxy.resultSet)
	return nil
}

func (um *sessionModule) register(scp *scope.Scope) {
	modScope := scope.New(scope.Opts{})
	mod := object.NewModule("session", modScope)

	modScope.AddBuiltins([]*object.Builtin{
		object.NewBuiltin("query", um.query, mod),
		object.NewBuiltin("result_set", um.resultSet, mod),
		object.NewBuiltin("selected_item", um.selectedItem, mod),
		object.NewBuiltin("set_result_set", um.setResultSet, mod),
	})

	scp.Declare("session", mod, true)
}
