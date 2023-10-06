package scriptmanager

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
	"github.com/risor-io/risor/object"
)

type sessionModule struct {
	sessionService SessionService
}

func (um *sessionModule) query(ctx context.Context, args ...object.Object) object.Object {
	if len(args) == 0 || len(args) > 2 {
		return object.Errorf("type error: session.query takes either 1 or 2 arguments (%d given)", len(args))
	}

	var options QueryOptions

	expr, objErr := object.AsString(args[0])
	if objErr != nil {
		return objErr
	}

	if len(args) == 2 {
		objMap, objErr := object.AsMap(args[1])
		if objErr != nil {
			return objErr
		}

		// Table name
		if val := objMap.Get("table"); val != object.Nil && val.IsTruthy() {
			switch tv := val.(type) {
			case *object.String:
				options.TableName = tv.Value()
			case *tableProxy:
				options.TableName = tv.table.Name
			default:
				return object.Errorf("type error: query option 'table' must be either a string or table")
			}
		}

		// Index name
		if val, isStr := objMap.Get("index").(*object.String); isStr {
			options.IndexName = val.Value()
		}

		// Placeholders
		if argsVal, isArgsValMap := objMap.Get("args").(*object.Map); isArgsValMap {
			options.NamePlaceholders = make(map[string]string)
			options.ValuePlaceholders = make(map[string]types.AttributeValue)

			for k, val := range argsVal.Value() {
				switch v := val.(type) {
				case *object.String:
					options.NamePlaceholders[k] = v.Value()
					options.ValuePlaceholders[k] = &types.AttributeValueMemberS{Value: v.Value()}
				case *object.Int:
					options.ValuePlaceholders[k] = &types.AttributeValueMemberN{Value: fmt.Sprint(v.Value())}
				case *object.Float:
					options.ValuePlaceholders[k] = &types.AttributeValueMemberN{Value: fmt.Sprint(v.Value())}
				case *object.Bool:
					options.ValuePlaceholders[k] = &types.AttributeValueMemberBOOL{Value: v.Value()}
				case *object.NilType:
					options.ValuePlaceholders[k] = &types.AttributeValueMemberNULL{Value: true}
				default:
					return object.Errorf("type error: arg '%v' of type '%v' is not supported", k, val.Type())
				}
			}
		}
	}

	resp, err := um.sessionService.Query(ctx, expr, options)

	if err != nil {
		return object.NewError(err)
	}
	return &resultSetProxy{resultSet: resp}
}

func (um *sessionModule) resultSet(ctx context.Context, args ...object.Object) object.Object {
	if err := require("session.result_set", 0, args); err != nil {
		return err
	}

	rs := um.sessionService.ResultSet(ctx)
	if rs == nil {
		return object.Nil
	}
	return &resultSetProxy{resultSet: rs}
}

func (um *sessionModule) selectedItem(ctx context.Context, args ...object.Object) object.Object {
	if err := require("session.result_set", 0, args); err != nil {
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
	if err := require("session.set_result_set", 1, args); err != nil {
		return err
	}

	resultSetProxy, isResultSetProxy := args[0].(*resultSetProxy)
	if !isResultSetProxy {
		return object.NewError(errors.Errorf("type error: expected a resultsset (got %v)", args[0]))
	}

	um.sessionService.SetResultSet(ctx, resultSetProxy.resultSet)
	return nil
}

func (um *sessionModule) currentTable(ctx context.Context, args ...object.Object) object.Object {
	if err := require("session.current_table", 0, args); err != nil {
		return err
	}

	rs := um.sessionService.ResultSet(ctx)
	if rs == nil {
		return object.Nil
	}

	return &tableProxy{table: rs.TableInfo}
}

func (um *sessionModule) register() *object.Module {
	return object.NewBuiltinsModule("session", map[string]object.Object{
		"query":          object.NewBuiltin("query", um.query),
		"current_table":  object.NewBuiltin("current_table", um.currentTable),
		"result_set":     object.NewBuiltin("result_set", um.resultSet),
		"selected_item":  object.NewBuiltin("selected_item", um.selectedItem),
		"set_result_set": object.NewBuiltin("set_result_set", um.setResultSet),
	})
}
