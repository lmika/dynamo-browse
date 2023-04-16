package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/common/maputils"
	"github.com/lmika/dynamo-browse/internal/common/sliceutils"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
)

type exprValue interface {
	typeName() string
	asGoValue() any
	asAttributeValue() types.AttributeValue
}

type stringableExprValue interface {
	exprValue
	asString() string
}

type numberableExprValue interface {
	exprValue
	asBigFloat() *big.Float
	asInt() int64
}

type slicableExprValue interface {
	exprValue
	len() int
	valueAt(idx int) (exprValue, error)
}

type modifiableSliceExprValue interface {
	setValueAt(idx int, value exprValue)
	deleteValueAt(idx int)
}

type mappableExprValue interface {
	len() int
	hasKey(name string) bool
	valueOf(name string) (exprValue, error)
}

type modifiableMapExprValue interface {
	setValueOf(name string, value exprValue)
	deleteValueOf(name string)
}

func buildExpressionFromValue(ev exprValue) expression.ValueBuilder {
	return expression.Value(ev)
}

func newExprValueFromAttributeValue(ev types.AttributeValue) (exprValue, error) {
	if ev == nil {
		return nil, nil
	}

	switch xVal := ev.(type) {
	case *types.AttributeValueMemberS:
		return stringExprValue(xVal.Value), nil
	case *types.AttributeValueMemberN:
		xNumVal, _, err := big.ParseFloat(xVal.Value, 10, 63, big.ToNearestEven)
		if err != nil {
			return nil, err
		}
		return bigNumExprValue{num: xNumVal}, nil
	case *types.AttributeValueMemberBOOL:
		return boolExprValue(xVal.Value), nil
	case *types.AttributeValueMemberNULL:
		return nullExprValue{}, nil
	case *types.AttributeValueMemberL:
		return listProxyValue{list: xVal}, nil
	case *types.AttributeValueMemberM:
		return mapProxyValue{mapValue: xVal}, nil
	case *types.AttributeValueMemberSS:
		return stringSetProxyValue{stringSet: xVal}, nil
	case *types.AttributeValueMemberNS:
		return numberSetProxyValue{numberSet: xVal}, nil
	}
	return nil, errors.New("cannot convert to expr value")
}

type stringExprValue string

func (s stringExprValue) asGoValue() any {
	return string(s)
}

func (s stringExprValue) asAttributeValue() types.AttributeValue {
	return &types.AttributeValueMemberS{Value: string(s)}
}

func (s stringExprValue) asString() string {
	return string(s)
}

func (s stringExprValue) typeName() string {
	return "S"
}

type int64ExprValue int64

func (i int64ExprValue) asGoValue() any {
	return int(i)
}

func (i int64ExprValue) asAttributeValue() types.AttributeValue {
	return &types.AttributeValueMemberN{Value: strconv.Itoa(int(i))}
}

func (i int64ExprValue) asInt() int64 {
	return int64(i)
}

func (i int64ExprValue) asBigFloat() *big.Float {
	var f big.Float
	f.SetInt64(int64(i))
	return &f
}

func (s int64ExprValue) typeName() string {
	return "N"
}

type bigNumExprValue struct {
	num *big.Float
}

func (i bigNumExprValue) asGoValue() any {
	return i.num
}

func (i bigNumExprValue) asAttributeValue() types.AttributeValue {
	return &types.AttributeValueMemberN{Value: i.num.String()}
}

func (i bigNumExprValue) asInt() int64 {
	x, _ := i.num.Int64()
	return x
}

func (i bigNumExprValue) asBigFloat() *big.Float {
	return i.num
}

func (s bigNumExprValue) typeName() string {
	return "N"
}

type boolExprValue bool

func (b boolExprValue) asGoValue() any {
	return bool(b)
}

func (b boolExprValue) asAttributeValue() types.AttributeValue {
	return &types.AttributeValueMemberBOOL{Value: bool(b)}
}

func (s boolExprValue) typeName() string {
	return "BOOL"
}

type nullExprValue struct{}

func (b nullExprValue) asGoValue() any {
	return nil
}

func (b nullExprValue) asAttributeValue() types.AttributeValue {
	return &types.AttributeValueMemberNULL{Value: true}
}

func (s nullExprValue) typeName() string {
	return "NULL"
}

type listExprValue []exprValue

func (bs listExprValue) asGoValue() any {
	return sliceutils.Map(bs, func(t exprValue) any {
		return t.asGoValue()
	})
}

func (bs listExprValue) asAttributeValue() types.AttributeValue {
	return &types.AttributeValueMemberL{Value: sliceutils.Map(bs, func(t exprValue) types.AttributeValue {
		return t.asAttributeValue()
	})}
}

func (bs listExprValue) len() int {
	return len(bs)
}

func (bs listExprValue) valueAt(i int) (exprValue, error) {
	return bs[i], nil
}

func (s listExprValue) typeName() string {
	return "L"
}

type mapExprValue map[string]exprValue

func (bs mapExprValue) asGoValue() any {
	return maputils.MapValues(bs, func(t exprValue) any {
		return t.asGoValue()
	})
}

func (bs mapExprValue) asAttributeValue() types.AttributeValue {
	return &types.AttributeValueMemberM{Value: maputils.MapValues(bs, func(t exprValue) types.AttributeValue {
		return t.asAttributeValue()
	})}
}

func (bs mapExprValue) len() int {
	return len(bs)
}

func (bs mapExprValue) hasKey(name string) bool {
	_, ok := bs[name]
	return ok
}

func (bs mapExprValue) valueOf(name string) (exprValue, error) {
	return bs[name], nil
}

func (s mapExprValue) typeName() string {
	return "M"
}

type listProxyValue struct {
	list *types.AttributeValueMemberL
}

func (bs listProxyValue) asGoValue() any {
	resultingList := make([]any, len(bs.list.Value))
	for i, item := range bs.list.Value {
		if av, _ := newExprValueFromAttributeValue(item); av != nil {
			resultingList[i] = av.asGoValue()
		} else {
			resultingList[i] = nil
		}
	}
	return resultingList
}

func (bs listProxyValue) asAttributeValue() types.AttributeValue {
	return bs.list
}

func (bs listProxyValue) len() int {
	return len(bs.list.Value)
}

func (bs listProxyValue) valueAt(i int) (exprValue, error) {
	return newExprValueFromAttributeValue(bs.list.Value[i])
}

func (bs listProxyValue) setValueAt(i int, newVal exprValue) {
	bs.list.Value[i] = newVal.asAttributeValue()
}

func (bs listProxyValue) deleteValueAt(idx int) {
	newList := append([]types.AttributeValue{}, bs.list.Value[:idx]...)
	newList = append(newList, bs.list.Value[idx+1:]...)
	bs.list = &types.AttributeValueMemberL{Value: newList}
}

func (s listProxyValue) typeName() string {
	return "L"
}

type mapProxyValue struct {
	mapValue *types.AttributeValueMemberM
}

func (bs mapProxyValue) asGoValue() any {
	resultingMap := make(map[string]any)
	for k, item := range bs.mapValue.Value {
		if av, _ := newExprValueFromAttributeValue(item); av != nil {
			resultingMap[k] = av.asGoValue()
		} else {
			resultingMap[k] = nil
		}
	}
	return resultingMap
}

func (bs mapProxyValue) asAttributeValue() types.AttributeValue {
	return bs.mapValue
}

func (bs mapProxyValue) len() int {
	return len(bs.mapValue.Value)
}

func (bs mapProxyValue) hasKey(name string) bool {
	_, ok := bs.mapValue.Value[name]
	return ok
}

func (bs mapProxyValue) valueOf(name string) (exprValue, error) {
	return newExprValueFromAttributeValue(bs.mapValue.Value[name])
}

func (bs mapProxyValue) setValueOf(name string, newVal exprValue) {
	bs.mapValue.Value[name] = newVal.asAttributeValue()
}

func (bs mapProxyValue) deleteValueOf(name string) {
	delete(bs.mapValue.Value, name)
}

func (s mapProxyValue) typeName() string {
	return "M"
}

type stringSetProxyValue struct {
	stringSet *types.AttributeValueMemberSS
}

func (bs stringSetProxyValue) asGoValue() any {
	return bs.stringSet.Value
}

func (bs stringSetProxyValue) asAttributeValue() types.AttributeValue {
	return bs.stringSet
}

func (bs stringSetProxyValue) len() int {
	return len(bs.stringSet.Value)
}

func (bs stringSetProxyValue) valueAt(i int) (exprValue, error) {
	return stringExprValue(bs.stringSet.Value[i]), nil
}

func (bs stringSetProxyValue) setValueAt(i int, newVal exprValue) {
	if str, isStr := newVal.(stringableExprValue); isStr {
		bs.stringSet.Value[i] = str.asString()
	}
}

func (bs stringSetProxyValue) deleteValueAt(idx int) {
	newList := append([]string{}, bs.stringSet.Value[:idx]...)
	newList = append(newList, bs.stringSet.Value[idx+1:]...)
	bs.stringSet = &types.AttributeValueMemberSS{Value: newList}
}

func (s stringSetProxyValue) typeName() string {
	return "SS"
}

type numberSetProxyValue struct {
	numberSet *types.AttributeValueMemberNS
}

func (bs numberSetProxyValue) asGoValue() any {
	return bs.numberSet.Value
}

func (bs numberSetProxyValue) asAttributeValue() types.AttributeValue {
	return bs.numberSet
}

func (bs numberSetProxyValue) len() int {
	return len(bs.numberSet.Value)
}

func (bs numberSetProxyValue) valueAt(i int) (exprValue, error) {
	fs, _, err := big.ParseFloat(bs.numberSet.Value[i], 10, 63, big.ToNearestEven)
	if err != nil {
		return nil, err
	}

	return bigNumExprValue{fs}, nil
}

func (bs numberSetProxyValue) setValueAt(i int, newVal exprValue) {
	if str, isStr := newVal.(numberableExprValue); isStr {
		bs.numberSet.Value[i] = str.asBigFloat().String()
	}
}

func (bs numberSetProxyValue) deleteValueAt(idx int) {
	newList := append([]string{}, bs.numberSet.Value[:idx]...)
	newList = append(newList, bs.numberSet.Value[idx+1:]...)
	bs.numberSet = &types.AttributeValueMemberNS{Value: newList}
}

func (s numberSetProxyValue) typeName() string {
	return "NS"
}
