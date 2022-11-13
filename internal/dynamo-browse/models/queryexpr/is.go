package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"strings"
)

type isTypeInfo struct {
	isAny         bool
	attributeType expression.DynamoDBAttributeType
}

var validIsTypeNames = map[string]isTypeInfo{
	"ANY":  {isAny: true},
	"B":    {attributeType: expression.Binary},
	"BOOL": {attributeType: expression.Boolean},
	"S":    {attributeType: expression.String},
	"N":    {attributeType: expression.Number},
	"NULL": {attributeType: expression.Null},
	"L":    {attributeType: expression.List},
	"M":    {attributeType: expression.Map},
	"BS":   {attributeType: expression.BinarySet},
	"NS":   {attributeType: expression.NumberSet},
	"SS":   {attributeType: expression.StringSet},
}

func (a *astIsOp) evalToIR(info *models.TableInfo) (irAtom, error) {
	leftIR, err := a.Ref.evalToIR(info)
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

	rightIR, err := a.Value.evalToIR(info)
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

func (a *astIsOp) evalItem(item models.Item) (types.AttributeValue, error) {
	ref, err := a.Ref.evalItem(item)
	if err != nil {
		return nil, err
	}

	if a.Value == nil {
		return ref, nil
	}
	panic("TODO")
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
