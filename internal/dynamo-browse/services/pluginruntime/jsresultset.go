package pluginruntime

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/dop251/goja"
	"github.com/lmika/audax/internal/dynamo-browse/models"
)

var resultSetSymbol = goja.NewSymbol("resultSetProxy")

func newJSResultSet(rt *goja.Runtime, resultSet *models.ResultSet) *goja.Object {
	obj := rt.NewObject()
	obj.DefineDataPropertySymbol(resultSetSymbol, rt.NewDynamicObject(goProxyValue{resultSet}), goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_FALSE)
	obj.DefineDataProperty("table", newJSTableInfo(rt, resultSet.TableInfo), goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_FALSE)
	obj.DefineDataProperty("rows", rt.NewDynamicArray(resultSetRowProxy{rt, resultSet}), goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_FALSE)
	return obj
}

func newJSTableInfo(rt *goja.Runtime, tableInfo *models.TableInfo) *goja.Object {
	obj := rt.NewObject()
	obj.DefineDataProperty("name", rt.ToValue(tableInfo.Name), goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_FALSE)
	return obj
}

func newItemInfo(rt *goja.Runtime, rs *models.ResultSet, idx int) *goja.Object {
	obj := rt.NewObject()
	obj.DefineDataProperty("item", rt.NewDynamicObject(resultSetItemProxy{rt, rs, idx}), goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_FALSE)
	return obj
}

type resultSetRowProxy struct {
	rt *goja.Runtime
	rs *models.ResultSet
}

func (r resultSetRowProxy) Len() int {
	return len(r.rs.Items())
}

func (r resultSetRowProxy) Get(idx int) goja.Value {
	return newItemInfo(r.rt, r.rs, idx)
}

func (r resultSetRowProxy) Set(idx int, val goja.Value) bool {
	return false
}

func (r resultSetRowProxy) SetLen(i int) bool {
	return false
}

type resultSetItemProxy struct {
	rt  *goja.Runtime
	rs  *models.ResultSet
	idx int
}

func (r resultSetItemProxy) Get(key string) goja.Value {
	switch v := r.rs.Items()[r.idx][key].(type) {
	case *types.AttributeValueMemberS:
		return r.rt.ToValue(v.Value)
	}
	return goja.Undefined()
}

func (r resultSetItemProxy) Set(key string, val goja.Value) bool {
	item := r.rs.Items()[r.idx]

	switch v := val.Export().(type) {
	case string:
		item[key] = &types.AttributeValueMemberS{Value: v}
		r.rs.SetDirty(r.idx, true)
		return true
	}
	return false
}

func (r resultSetItemProxy) Has(key string) bool {
	_, hasKey := r.rs.Items()[r.idx][key]
	return hasKey
}

func (r resultSetItemProxy) Delete(key string) bool {
	item := r.rs.Items()[r.idx]
	delete(item, key)

	r.rs.SetDirty(r.idx, true)
	return true
}

func (r resultSetItemProxy) Keys() []string {
	keys := make([]string, 0, len(r.rs.Items()[r.idx]))
	for k := range r.rs.Items()[r.idx] {
		keys = append(keys, k)
	}
	return keys
}

type goProxyValue struct{ v any }

func (g goProxyValue) Get(key string) goja.Value {
	return goja.Undefined()
}

func (g goProxyValue) Set(key string, val goja.Value) bool {
	return false
}

func (g goProxyValue) Has(key string) bool {
	return false
}

func (g goProxyValue) Delete(key string) bool {
	return false
}

func (g goProxyValue) Keys() []string {
	return nil
}
