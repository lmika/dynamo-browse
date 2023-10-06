package scriptmanager

import (
	"github.com/lmika/dynamo-browse/internal/common/sliceutils"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"github.com/risor-io/risor/object"
	"github.com/risor-io/risor/op"
	"reflect"
)

const (
	tableProxyPartitionKey = "hash"
	tableProxySortKey      = "range"
)

type tableProxy struct {
	table *models.TableInfo
}

func (t *tableProxy) SetAttr(name string, value object.Object) error {
	return errors.Errorf("attribute error: %v", name)
}

func (t *tableProxy) RunOperation(opType op.BinaryOpType, right object.Object) object.Object {
	return object.Errorf("op error: unsupported %v", opType)
}

func (t *tableProxy) Cost() int {
	return 0
}

func (t *tableProxy) Type() object.Type {
	return "table"
}

func (t *tableProxy) Inspect() string {
	return "table(" + t.table.Name + ")"
}

func (t *tableProxy) Interface() interface{} {
	return t.table
}

func (t *tableProxy) Equals(other object.Object) object.Object {
	otherT, isOtherRS := other.(*tableProxy)
	if !isOtherRS {
		return object.False
	}

	return object.NewBool(reflect.DeepEqual(t.table, otherT.table))
}

func (t *tableProxy) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "name":
		return object.NewString(t.table.Name), true
	case "keys":
		if t.table.Keys.SortKey == "" {
			return object.NewMap(map[string]object.Object{
				tableProxyPartitionKey: object.NewString(t.table.Keys.PartitionKey),
				tableProxySortKey:      object.Nil,
			}), true
		}

		return object.NewMap(map[string]object.Object{
			tableProxyPartitionKey: object.NewString(t.table.Keys.PartitionKey),
			tableProxySortKey:      object.NewString(t.table.Keys.SortKey),
		}), true
	case "gsis":
		return object.NewList(sliceutils.Map(t.table.GSIs, newTableIndexProxy)), true
	}

	return nil, false
}

func (t *tableProxy) IsTruthy() bool {
	return true
}

type tableIndexProxy struct {
	gsi models.TableGSI
}

func (t tableIndexProxy) SetAttr(name string, value object.Object) error {
	return errors.Errorf("attribute error: %v", name)
}

func (t tableIndexProxy) RunOperation(opType op.BinaryOpType, right object.Object) object.Object {
	return object.Errorf("op error: unsupported %v", opType)
}

func (t tableIndexProxy) Cost() int {
	return 0
}

func newTableIndexProxy(gsi models.TableGSI) object.Object {
	return tableIndexProxy{gsi: gsi}
}

func (t tableIndexProxy) Type() object.Type {
	return "table_index"
}

func (t tableIndexProxy) Inspect() string {
	return "table_index(gsi," + t.gsi.Name + ")"
}

func (t tableIndexProxy) Interface() interface{} {
	return t.gsi
}

func (t tableIndexProxy) Equals(other object.Object) object.Object {
	otherIP, isOtherIP := other.(tableIndexProxy)
	if !isOtherIP {
		return object.False
	}

	return object.NewBool(reflect.DeepEqual(t.gsi, otherIP.gsi))
}

func (t tableIndexProxy) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "name":
		return object.NewString(t.gsi.Name), true
	case "keys":
		return object.NewMap(map[string]object.Object{
			tableProxyPartitionKey: object.NewString(t.gsi.Keys.PartitionKey),
			tableProxySortKey:      object.NewString(t.gsi.Keys.SortKey),
		}), true
	}

	return nil, false
}

func (t tableIndexProxy) IsTruthy() bool {
	return true
}
