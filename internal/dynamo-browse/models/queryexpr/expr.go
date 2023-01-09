package queryexpr

import (
	"bytes"
	"encoding/gob"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/attrcodec"
	"github.com/lmika/audax/internal/dynamo-browse/models/attrutils"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"hash/fnv"
	"io"
)

type QueryExpr struct {
	ast    *astExpr
	names  map[string]string
	values map[string]types.AttributeValue
}

type serializedExpr struct {
	Expr   string
	Names  map[string]string
	Values []byte
}

func DeserializeFrom(r io.Reader) (*QueryExpr, error) {
	var se serializedExpr

	if err := gob.NewDecoder(r).Decode(&se); err != nil {
		return nil, err
	}

	qe, err := Parse(se.Expr)
	if err != nil {
		return nil, err
	}

	qe.names = se.Names

	if len(se.Values) > 0 {
		vals, err := attrcodec.NewDecoder(bytes.NewReader(se.Values)).Decode()
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal placeholder values")
		}
		mvals, ok := vals.(*types.AttributeValueMemberM)
		if !ok {
			return nil, errors.Errorf("expected marshaled placeholder values to be map, but was %T", vals)
		}
		qe.values = mvals.Value
	}

	return qe, nil
}

func (md *QueryExpr) SerializeTo(w io.Writer) error {
	se := serializedExpr{Expr: md.String(), Names: md.names}
	if md.values != nil {
		var bts bytes.Buffer
		if err := attrcodec.NewEncoder(&bts).Encode(&types.AttributeValueMemberM{Value: md.values}); err != nil {
			return errors.Wrap(err, "unable to unmarshal placeholder values")
		}
		se.Values = bts.Bytes()
	}

	return gob.NewEncoder(w).Encode(se)
}

func (md *QueryExpr) SerializeToBytes() ([]byte, error) {
	if md == nil {
		return nil, nil
	}
	var bfr bytes.Buffer

	if err := md.SerializeTo(&bfr); err != nil {
		return nil, err
	}
	return bfr.Bytes(), nil
}

// Equal returns true if a query expression is equal another one.  Two query expressions are equal if they
// have the same query and placeholder values.  This is resistant to map ordering.
func (md *QueryExpr) Equal(other *QueryExpr) bool {
	if md == nil {
		return other == nil
	}
	return md.ast.String() == other.ast.String() &&
		maps.Equal(md.names, other.names) &&
		maps.EqualFunc(md.values, md.values, attrutils.Equals)
}

// HashCode will return a hash-code for this query expression.  This is to assist with determine whether two
// queries are the same.  If two queries have the same hash code, they may be equals (this will need to be
// confirmed by calling Equal()).  Otherwise, the queries cannot be equals.
func (md *QueryExpr) HashCode() uint64 {
	if md == nil {
		return 0
	}

	h := fnv.New64a()
	h.Write([]byte(md.ast.String()))

	// the names must be in sorted order to maintain consistant key ordering
	if len(md.names) > 0 {
		sortedKeys := make([]string, len(md.names))
		copy(sortedKeys, maps.Keys(md.names))
		slices.Sort(sortedKeys)

		for _, k := range sortedKeys {
			h.Write([]byte(k))
			h.Write([]byte(md.names[k]))
		}
	}

	if len(md.values) > 0 {
		sortedKeys := make([]string, len(md.values))
		copy(sortedKeys, maps.Keys(md.values))
		slices.Sort(sortedKeys)

		for _, k := range sortedKeys {
			h.Write([]byte(k))
			attrutils.HashTo(h, md.values[k])
		}
	}

	return h.Sum64()
}

func (md *QueryExpr) WithNameParams(value map[string]string) *QueryExpr {
	return &QueryExpr{
		ast:    md.ast,
		names:  value,
		values: md.values,
	}
}

func (md *QueryExpr) NameParam(name string) (string, bool) {
	return md.evalContext().lookupName(name)
}

func (md *QueryExpr) ValueParam(name string) (types.AttributeValue, bool) {
	return md.evalContext().lookupValue(name)
}

func (md *QueryExpr) ValueParamOrNil(name string) types.AttributeValue {
	v, ok := md.ValueParam(name)
	if !ok {
		return nil
	}
	return v
}

func (md *QueryExpr) WithValueParams(value map[string]types.AttributeValue) *QueryExpr {
	return &QueryExpr{
		ast:    md.ast,
		names:  md.names,
		values: value,
	}
}

func (md *QueryExpr) Plan(tableInfo *models.TableInfo) (*models.QueryExecutionPlan, error) {
	return md.ast.calcQuery(md.evalContext(), tableInfo)
}

func (md *QueryExpr) EvalItem(item models.Item) (types.AttributeValue, error) {
	return md.ast.evalItem(md.evalContext(), item)
}

func (md *QueryExpr) DeleteAttribute(item models.Item) error {
	return md.ast.deleteAttribute(md.evalContext(), item)
}

func (md *QueryExpr) SetEvalItem(item models.Item, newValue types.AttributeValue) error {
	return md.ast.setEvalItem(md.evalContext(), item, newValue)
}

func (md *QueryExpr) IsModifiablePath(item models.Item) bool {
	return md.ast.canModifyItem(md.evalContext(), item)
}

func (md *QueryExpr) evalContext() *evalContext {
	return &evalContext{
		namePlaceholders:  md.names,
		valuePlaceholders: md.values,
	}
}

func (md *QueryExpr) String() string {
	return md.ast.String()
}

func (a *astExpr) String() string {
	return a.Root.String()
}

type queryCalcInfo struct {
	seenKeys map[string]struct{}
}

func (qc *queryCalcInfo) clone() *queryCalcInfo {
	newKeys := make(map[string]struct{})
	for k, v := range qc.seenKeys {
		newKeys[k] = v
	}
	return &queryCalcInfo{seenKeys: newKeys}
}

func (qc *queryCalcInfo) hasSeenPrimaryKey(tableInfo *models.TableInfo) bool {
	_, hasKey := qc.seenKeys[tableInfo.Keys.PartitionKey]
	return hasKey
}

func (qc *queryCalcInfo) addKey(tableInfo *models.TableInfo, key string) bool {
	if tableInfo.Keys.PartitionKey != key && tableInfo.Keys.SortKey != key {
		return false
	}

	if qc.seenKeys == nil {
		qc.seenKeys = make(map[string]struct{})
	}
	if _, hasSeenKey := qc.seenKeys[key]; hasSeenKey {
		return false
	}

	qc.seenKeys[key] = struct{}{}
	return true
}

type evalContext struct {
	namePlaceholders  map[string]string
	nameLookup        func(string) (string, bool)
	valuePlaceholders map[string]types.AttributeValue
	valueLookup       func(string) (types.AttributeValue, bool)
}

func (ec *evalContext) lookupName(name string) (string, bool) {
	val, hasVal := ec.namePlaceholders[name]
	if hasVal {
		return val, true
	}

	if fn := ec.nameLookup; fn != nil {
		return fn(name)
	}

	return "", false
}

func (ec *evalContext) lookupValue(name string) (types.AttributeValue, bool) {
	val, hasVal := ec.valuePlaceholders[name]
	if hasVal {
		return val, true
	}

	if fn := ec.valueLookup; fn != nil {
		return fn(name)
	}

	return nil, false
}
