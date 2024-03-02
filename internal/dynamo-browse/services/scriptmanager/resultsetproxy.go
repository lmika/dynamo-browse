package scriptmanager

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/attrutils"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/queryexpr"
	"github.com/pkg/errors"
	"github.com/risor-io/risor/object"
	"github.com/risor-io/risor/op"
)

type resultSetProxy struct {
	resultSet *models.ResultSet
}

func newResultSetProxy(rs *models.ResultSet) *resultSetProxy {
	return &resultSetProxy{resultSet: rs}
}

func (r *resultSetProxy) SetAttr(name string, value object.Object) error {
	return errors.Errorf("attribute error: %v", name)
}

func (r *resultSetProxy) RunOperation(opType op.BinaryOpType, right object.Object) object.Object {
	return object.Errorf("op error: unsupported %v", opType)
}

func (r *resultSetProxy) Cost() int {
	return len(r.resultSet.Items())
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
	otherRS, isOtherRS := other.(*resultSetProxy)
	if !isOtherRS {
		return object.False
	}

	return object.NewBool(r.resultSet == otherRS.resultSet)
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

	return newItemProxy(r, realIdx), nil
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
	case "table":
		return &tableProxy{table: r.resultSet.TableInfo}, true
	case "length":
		return object.NewInt(int64(len(r.resultSet.Items()))), true
	case "find":
		return object.NewBuiltin("find", r.find), true
	case "merge":
		return object.NewBuiltin("merge", r.merge), true
	}

	return nil, false
}

func (i *resultSetProxy) find(ctx context.Context, args ...object.Object) object.Object {
	if objErr := require("resultset.find", 1, args); objErr != nil {
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

	for idx, item := range i.resultSet.Items() {
		rs, err := modExpr.EvalItem(item)
		if err != nil {
			continue
		}

		if attrutils.Truthy(rs) {
			return newItemProxy(i, idx)
		}
	}

	return object.Nil
}

func (i *resultSetProxy) merge(ctx context.Context, args ...object.Object) object.Object {
	type pksk struct {
		pk types.AttributeValue
		sk types.AttributeValue
	}

	if objErr := require("resultset.merge", 1, args); objErr != nil {
		return objErr
	}

	otherRS, isRS := args[0].(*resultSetProxy)
	if !isRS {
		return object.NewError(errors.Errorf("type error: expected a resultset (got %v)", args[0].Type()))
	}

	if !i.resultSet.TableInfo.Equal(otherRS.resultSet.TableInfo) {
		return object.Nil
	}

	itemsInI := make(map[pksk]models.Item)
	newItems := make([]models.Item, 0, len(i.resultSet.Items())+len(otherRS.resultSet.Items()))
	for _, item := range i.resultSet.Items() {
		pk, sk := item.PKSK(i.resultSet.TableInfo)
		itemsInI[pksk{pk, sk}] = item
		newItems = append(newItems, item)
	}

	for _, item := range otherRS.resultSet.Items() {
		pk, sk := item.PKSK(i.resultSet.TableInfo)
		if _, hasItem := itemsInI[pksk{pk, sk}]; !hasItem {
			newItems = append(newItems, item)
		}
	}

	newResultSet := &models.ResultSet{
		Created:   time.Now(),
		TableInfo: i.resultSet.TableInfo,
	}
	newResultSet.SetItems(newItems)

	return &resultSetProxy{resultSet: newResultSet}
}

type itemProxy struct {
	resultSetProxy *resultSetProxy
	itemIndex      int
	item           models.Item
}

func (i *itemProxy) SetAttr(name string, value object.Object) error {
	return errors.Errorf("attribute error: %v", name)
}

func (i *itemProxy) RunOperation(opType op.BinaryOpType, right object.Object) object.Object {
	return object.Errorf("op error: unsupported %v", opType)
}

func (i *itemProxy) Cost() int {
	return len(i.item)
}

func newItemProxy(rs *resultSetProxy, itemIndex int) *itemProxy {
	return &itemProxy{
		resultSetProxy: rs,
		itemIndex:      itemIndex,
		item:           rs.resultSet.Items()[itemIndex],
	}
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
	case "result_set":
		return i.resultSetProxy, true
	case "index":
		return object.NewInt(int64(i.itemIndex)), true
	case "attr":
		return object.NewBuiltin("attr", i.value), true
	case "set_attr":
		return object.NewBuiltin("set_attr", i.setValue), true
	case "delete_attr":
		return object.NewBuiltin("delete_attr", i.deleteAttr), true
	}

	return nil, false
}

func (i *itemProxy) value(ctx context.Context, args ...object.Object) object.Object {
	if objErr := require("item.attr", 1, args); objErr != nil {
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

	tVal, err := attributeValueToTamarin(av)
	if err != nil {
		return object.NewError(err)
	}
	return tVal
}

func (i *itemProxy) setValue(ctx context.Context, args ...object.Object) object.Object {
	if objErr := require("item.set_attr", 2, args); objErr != nil {
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

	newValue, err := tamarinValueToAttributeValue(args[1])
	if err != nil {
		return object.NewError(err)
	}
	if err := path.SetEvalItem(i.item, newValue); err != nil {
		return object.NewError(err)
	}

	i.resultSetProxy.resultSet.SetDirty(i.itemIndex, true)
	return nil
}

func (i *itemProxy) deleteAttr(ctx context.Context, args ...object.Object) object.Object {
	if objErr := require("item.delete_attr", 1, args); objErr != nil {
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
	if err := modExpr.DeleteAttribute(i.item); err != nil {
		return object.NewError(errors.Errorf("arg error: path expression evaluate error: %v", err))
	}

	i.resultSetProxy.resultSet.SetDirty(i.itemIndex, true)
	return nil
}
