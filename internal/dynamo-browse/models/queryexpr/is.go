package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/attrutils"
	"reflect"
	"strings"
)

type isTypeInfo struct {
	isAny         bool
	attributeType expression.DynamoDBAttributeType
	goType        reflect.Type
}

var validIsTypeNames = map[string]isTypeInfo{
	"ANY": {isAny: true},
	"B": {
		attributeType: expression.Binary,
		goType:        reflect.TypeOf(&types.AttributeValueMemberB{}),
	},
	"BOOL": {
		attributeType: expression.Boolean,
		goType:        reflect.TypeOf(&types.AttributeValueMemberBOOL{}),
	},
	"S": {
		attributeType: expression.String,
		goType:        reflect.TypeOf(&types.AttributeValueMemberS{}),
	},
	"N": {
		attributeType: expression.Number,
		goType:        reflect.TypeOf(&types.AttributeValueMemberN{}),
	},
	"NULL": {
		attributeType: expression.Null,
		goType:        reflect.TypeOf(&types.AttributeValueMemberNULL{}),
	},
	"L": {
		attributeType: expression.List,
		goType:        reflect.TypeOf(&types.AttributeValueMemberL{}),
	},
	"M": {
		attributeType: expression.Map,
		goType:        reflect.TypeOf(&types.AttributeValueMemberM{}),
	},
	"BS": {
		attributeType: expression.BinarySet,
		goType:        reflect.TypeOf(&types.AttributeValueMemberBS{}),
	},
	"NS": {
		attributeType: expression.NumberSet,
		goType:        reflect.TypeOf(&types.AttributeValueMemberNS{}),
	},
	"SS": {
		attributeType: expression.StringSet,
		goType:        reflect.TypeOf(&types.AttributeValueMemberSS{}),
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
	strValue, isStringValue := valueIR.goValue().(string)
	if !isStringValue {
		return nil, ValueMustBeStringError{}
	}

	typeInfo, isValidType := validIsTypeNames[strings.ToUpper(strValue)]
	if !isValidType {
		return nil, InvalidTypeForIsError{TypeName: strValue}
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

func (a *astIsOp) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
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
	str, canToStr := attrutils.AttributeToString(expTypeVal)
	if !canToStr {
		return nil, ValueMustBeStringError{}
	}
	typeInfo, hasTypeInfo := validIsTypeNames[strings.ToUpper(str)]
	if !hasTypeInfo {
		return nil, InvalidTypeForIsError{TypeName: str}
	}

	var resultOfIs bool
	if typeInfo.isAny {
		resultOfIs = ref != nil
	} else {
		refType := reflect.TypeOf(ref)
		resultOfIs = typeInfo.goType.AssignableTo(refType)
	}
	if a.HasNot {
		resultOfIs = !resultOfIs
	}
	return &types.AttributeValueMemberBOOL{Value: resultOfIs}, nil
}

func (a *astIsOp) canModifyItem(ctx *evalContext, item models.Item) bool {
	if a.Value != nil {
		return false
	}
	return a.Ref.canModifyItem(ctx, item)
}

func (a *astIsOp) setEvalItem(ctx *evalContext, item models.Item, value types.AttributeValue) error {
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
