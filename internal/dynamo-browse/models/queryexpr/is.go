package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"reflect"
	"strings"
)

type isTypeInfo struct {
	isAny         bool
	attributeType expression.DynamoDBAttributeType
	goTypes       []reflect.Type
}

var validIsTypeNames = map[string]isTypeInfo{
	"ANY": {isAny: true},
	"B": {
		attributeType: expression.Binary,
		// TODO
	},
	"BOOL": {
		attributeType: expression.Boolean,
		goTypes:       []reflect.Type{reflect.TypeOf(boolExprValue(false))},
	},
	"S": {
		attributeType: expression.String,
		goTypes:       []reflect.Type{reflect.TypeOf(stringExprValue(""))},
	},
	"N": {
		attributeType: expression.Number,
		goTypes:       []reflect.Type{reflect.TypeOf(int64ExprValue(0)), reflect.TypeOf(bigNumExprValue{})},
	},
	"NULL": {
		attributeType: expression.Null,
		goTypes:       []reflect.Type{reflect.TypeOf(nullExprValue{})},
	},
	"L": {
		attributeType: expression.List,
		goTypes:       []reflect.Type{reflect.TypeOf(listExprValue{}), reflect.TypeOf(listProxyValue{})},
	},
	"M": {
		attributeType: expression.Map,
		goTypes:       []reflect.Type{reflect.TypeOf(mapExprValue{}), reflect.TypeOf(mapProxyValue{})},
	},
	"BS": {
		attributeType: expression.BinarySet,
		// TODO
	},
	"NS": {
		attributeType: expression.NumberSet,
		// TODO
	},
	"SS": {
		attributeType: expression.StringSet,
		// TODO
	},
}

func (a *astIsOp) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	leftIR, err := a.Ref.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}

	if a.Value == nil {
		return leftIR, nil
	}

	nameIR, isNameIR := leftIR.(irNamePath)
	if !isNameIR {
		return nil, OperandNotANameError(a.Ref.String())
	}

	rightIR, err := a.Value.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}

	valueIR, isValueIR := rightIR.(irValue)
	if !isValueIR {
		return nil, ValueMustBeLiteralError{}
	}
	strValue, isStringValue := valueIR.exprValue().(stringableExprValue)
	if !isStringValue {
		return nil, ValueMustBeStringError{}
	}

	typeInfo, isValidType := validIsTypeNames[strings.ToUpper(strValue.asString())]
	if !isValidType {
		return nil, InvalidTypeForIsError{TypeName: strValue.asString()}
	}

	var ir = irIs{name: nameIR, typeInfo: typeInfo}
	if a.HasNot {
		if typeInfo.isAny {
			ir.hasNot = true
		} else {
			return &irBoolNot{atom: ir}, nil
		}
	}
	return ir, nil
}

func (a *astIsOp) evalItem(ctx *evalContext, item models.Item) (exprValue, error) {
	ref, err := a.Ref.evalItem(ctx, item)
	if err != nil {
		return nil, err
	}

	if a.Value == nil {
		return ref, nil
	}

	expTypeVal, err := a.Value.evalItem(ctx, item)
	if err != nil {
		return nil, err
	}
	str, canToStr := expTypeVal.(stringableExprValue)
	if !canToStr {
		return nil, ValueMustBeStringError{}
	}
	typeInfo, hasTypeInfo := validIsTypeNames[strings.ToUpper(str.asString())]
	if !hasTypeInfo {
		return nil, InvalidTypeForIsError{TypeName: str.asString()}
	}

	var resultOfIs bool
	if typeInfo.isAny {
		resultOfIs = ref != nil
	} else {
		refType := reflect.TypeOf(ref)

		for _, t := range typeInfo.goTypes {
			if t.AssignableTo(refType) {
				resultOfIs = true
				break
			}
		}
	}
	if a.HasNot {
		resultOfIs = !resultOfIs
	}
	return boolExprValue(resultOfIs), nil
}

func (a *astIsOp) canModifyItem(ctx *evalContext, item models.Item) bool {
	if a.Value != nil {
		return false
	}
	return a.Ref.canModifyItem(ctx, item)
}

func (a *astIsOp) setEvalItem(ctx *evalContext, item models.Item, value exprValue) error {
	if a.Value != nil {
		return PathNotSettableError{}
	}
	return a.Ref.setEvalItem(ctx, item, value)
}

func (a *astIsOp) deleteAttribute(ctx *evalContext, item models.Item) error {
	if a.Value != nil {
		return PathNotSettableError{}
	}
	return a.Ref.deleteAttribute(ctx, item)
}

func (a *astIsOp) String() string {
	var sb strings.Builder

	sb.WriteString(a.Ref.String())
	if a.Value != nil {
		sb.WriteString(" is ")
		if a.HasNot {
			sb.WriteString("not ")
		}
		sb.WriteString(a.Value.String())
	}
	return sb.String()
}

type irIs struct {
	name     nameIRAtom
	hasNot   bool
	typeInfo isTypeInfo
}

func (i irIs) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	nb := i.name.calcName(info)
	if i.typeInfo.isAny {
		if i.hasNot {
			return expression.AttributeNotExists(nb), nil
		}
		return expression.AttributeExists(nb), nil
	}
	return expression.AttributeType(nb, i.typeInfo.attributeType), nil
}
