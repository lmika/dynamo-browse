package scriptmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cloudcmds/tamarin/arg"
	"github.com/cloudcmds/tamarin/object"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/queryexpr"
	"github.com/pkg/errors"
	"strconv"
)

type resultSetProxy struct {
	resultSet *models.ResultSet
}

func (r *resultSetProxy) Interface() interface{} {
	return r.resultSet
}

func (r *resultSetProxy) IsTruthy() bool {
	return true
}

func (r *resultSetProxy) Type() object.Type {
	return "resultset"
}

func (r *resultSetProxy) Inspect() string {
	return "resultset"
}

func (r *resultSetProxy) Equals(other object.Object) object.Object {
	// TODO
	return object.False
}

// GetItem implements the [key] operator for a container type.
func (r *resultSetProxy) GetItem(key object.Object) (object.Object, *object.Error) {
	idx, err := object.AsInt(key)
	if err != nil {
		return nil, err
	}

	realIdx := int(idx)
	if realIdx < 0 {
		realIdx = len(r.resultSet.Items()) + realIdx
	}

	if realIdx < 0 || realIdx >= len(r.resultSet.Items()) {
		return nil, object.NewError(errors.Errorf("index error: index out of range: %v", idx))
	}

	return &itemProxy{
		resultSet: r.resultSet,
		itemIndex: realIdx,
		item:      r.resultSet.Items()[realIdx],
	}, nil
}

// GetSlice implements the [start:stop] operator for a container type.
func (r *resultSetProxy) GetSlice(s object.Slice) (object.Object, *object.Error) {
	return nil, object.NewError(errors.New("TODO"))
}

// SetItem implements the [key] = value operator for a container type.
func (r *resultSetProxy) SetItem(key, value object.Object) *object.Error {
	return object.NewError(errors.New("TODO"))
}

// DelItem implements the del [key] operator for a container type.
func (r *resultSetProxy) DelItem(key object.Object) *object.Error {
	return object.NewError(errors.New("TODO"))
}

// Contains returns true if the given item is found in this container.
func (r *resultSetProxy) Contains(item object.Object) *object.Bool {
	// TODO
	return object.False
}

// Len returns the number of items in this container.
func (r *resultSetProxy) Len() *object.Int {
	return object.NewInt(int64(len(r.resultSet.Items())))
}

// Iter returns an iterator for this container.
func (r *resultSetProxy) Iter() object.Iterator {
	// TODO
	return nil
}

func (r *resultSetProxy) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "length":
		return object.NewInt(int64(len(r.resultSet.Items()))), true
	case "at":
		return object.NewBuiltin("at", r.at), true
	}

	return nil, false
}

func (r *resultSetProxy) at(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("resultset.at", 1, args); err != nil {
		return err
	}

	idx, err := object.AsInt(args[0])
	if err != nil {
		return err
	}

	realIdx := int(idx)
	if realIdx < 0 {
		realIdx = len(r.resultSet.Items()) + realIdx
	}

	if realIdx < 0 || realIdx >= len(r.resultSet.Items()) {
		return object.NewError(errors.Errorf("index error: index out of range: %v", idx))
	}

	return &itemProxy{
		resultSet: r.resultSet,
		itemIndex: realIdx,
		item:      r.resultSet.Items()[realIdx],
	}
}

type itemProxy struct {
	resultSet *models.ResultSet
	itemIndex int
	item      models.Item
}

func (i *itemProxy) Interface() interface{} {
	return i.item
}

func (i *itemProxy) IsTruthy() bool {
	return true
}

func (i *itemProxy) Type() object.Type {
	return "item"
}

func (i *itemProxy) Inspect() string {
	return "item"
}

func (i *itemProxy) Equals(other object.Object) object.Object {
	// TODO
	return object.False
}

func (i *itemProxy) GetAttr(name string) (object.Object, bool) {
	// TODO: this should implement the container interface
	switch name {
	case "value":
		return object.NewBuiltin("value", i.value), true
	case "set_value":
		return object.NewBuiltin("set_value", i.setValue), true
	}

	return nil, false
}

func (i *itemProxy) value(ctx context.Context, args ...object.Object) object.Object {
	if objErr := arg.Require("item.value", 1, args); objErr != nil {
		return objErr
	}

	str, objErr := object.AsString(args[0])
	if objErr != nil {
		return objErr
	}

	modExpr, err := queryexpr.Parse(str)
	if err != nil {
		return object.Errorf("arg error: invalid path expression: %v", err)
	}
	av, err := modExpr.EvalItem(i.item)
	if err != nil {
		return object.NewError(errors.Errorf("arg error: path expression evaluate error: %v", err))
	}

	// TODO
	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		return object.NewString(v.Value)
	case *types.AttributeValueMemberN:
		// TODO: better
		f, err := strconv.ParseFloat(v.Value, 64)
		if err != nil {
			return object.NewError(errors.Errorf("value error: invalid N value: %v", v.Value))
		}
		return object.NewFloat(f)
	}
	return object.NewError(errors.New("TODO"))
}

func (i *itemProxy) setValue(ctx context.Context, args ...object.Object) object.Object {
	if objErr := arg.Require("item.set_value", 2, args); objErr != nil {
		return objErr
	}

	pathExpr, objErr := object.AsString(args[0])
	if objErr != nil {
		return objErr
	}

	path, err := queryexpr.Parse(pathExpr)
	if err != nil {
		return object.Errorf("arg error: invalid path expression: %v", err)
	}

	//modExpr, err := queryexpr.Parse(str)
	//if err != nil {
	//	return object.NewError(errors.Errorf("arg error: invalid path expression: %v", err))
	//}
	//av, err := modExpr.EvalItem(i.item)
	//if err != nil {
	//	return object.NewError(errors.Errorf("arg error: path expression evaluate error: %v", err))
	//}

	// TODO
	newValue := args[1]
	switch v := newValue.(type) {
	case *object.String:
		if err := path.SetEvalItem(i.item, &types.AttributeValueMemberS{Value: v.Value()}); err != nil {
			return object.NewError(err)
		}
	//case *types.AttributeValueMemberN:
	//	// TODO: better
	//	f, err := strconv.ParseFloat(v.Value, 64)
	//	if err != nil {
	//		return object.NewError(errors.Errorf("value error: invalid N value: %v", v.Value))
	//	}
	//	return object.NewFloat(f)
	default:
		return object.Errorf("type error: unsupported value type (got %v)", newValue.Type())
	}

	i.resultSet.SetDirty(i.itemIndex, true)
	return nil
}
