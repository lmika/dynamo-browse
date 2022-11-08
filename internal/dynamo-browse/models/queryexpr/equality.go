package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/attrutils"
	"github.com/pkg/errors"
	"strings"
)

func (a *astEqualityOp) evalToIR(info *models.TableInfo) (irAtom, error) {
	if a.Op == "" {
		return a.Ref.evalToIR(info)
	}

	v, err := a.Value.rightOperandGoValue()
	if err != nil {
		return nil, err
	}

	singleName, isSingleName := a.Ref.unqualifiedName()
	if !isSingleName {
		return nil, errors.Errorf("%v: cannot use dereferences", singleName)
	}

	switch a.Op {
	case "=":
		return irFieldEq{name: singleName, value: v}, nil
	case "!=":
		return irFieldNe{name: singleName, value: v}, nil
	case "^=":
		strValue, isStrValue := v.(string)
		if !isStrValue {
			return nil, errors.New("operand '^=' must be string")
		}
		return irFieldBeginsWith{name: singleName, prefix: strValue}, nil
	}

	return nil, errors.Errorf("unrecognised operator: %v", a.Op)
}

func (a *astEqualityOp) rightOperandGoValue() (any, error) {
	if a.Op == "" {
		return a.Ref.rightOperandGoValue()
	}
	return nil, ValueMustBeLiteral{}
}

func (a *astEqualityOp) leftOperandName() (string, bool) {
	return a.Ref.unqualifiedName()
}

func (a *astEqualityOp) evalItem(item models.Item) (types.AttributeValue, error) {
	left, err := a.Ref.evalItem(item)
	if err != nil {
		return nil, err
	}

	if a.Op == "" {
		return left, nil
	}

	right, err := a.Value.rightOperandDynamoValue()
	if err != nil {
		return nil, err
	}

	switch a.Op {
	case "=":
		cmp, isComparable := attrutils.CompareScalarAttributes(left, right)
		if !isComparable {
			return nil, ValuesNotComparable{Left: left, Right: right}
		}
		return &types.AttributeValueMemberBOOL{Value: cmp == 0}, nil
	case "^=":
		rightVal, err := a.Value.rightOperandGoValue()
		if err != nil {
			return nil, err
		}

		strValue, isStrValue := rightVal.(string)
		if !isStrValue {
			return nil, errors.New("operand '^=' must be string")
		}

		leftAsStr, canBeString := attrutils.AttributeToString(left)
		if !canBeString {
			return nil, ValueNotConvertableToString{Val: left}
		}
		return &types.AttributeValueMemberBOOL{Value: strings.HasPrefix(leftAsStr, strValue)}, nil
	}

	return nil, errors.Errorf("unrecognised operator: %v", a.Op)
}

func (a *astEqualityOp) String() string {
	return a.Ref.String() + a.Op + a.Value.String()
}

type irFieldEq struct {
	name  string
	value any
}

func (a irFieldEq) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	if a.name == info.Keys.PartitionKey || a.name == info.Keys.SortKey {
		return qci.addKey(info, a.name)
	}

	return false
}

func (a irFieldEq) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	return expression.Name(a.name).Equal(expression.Value(a.value)), nil
}

func (a irFieldEq) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	return expression.Key(a.name).Equal(expression.Value(a.value)), nil
}

func (a irFieldEq) operandFieldName() string {
	return a.name
}

type irFieldNe struct {
	name  string
	value any
}

func (a irFieldNe) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	return false
}

func (a irFieldNe) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	return expression.Name(a.name).NotEqual(expression.Value(a.value)), nil
}

func (a irFieldNe) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	return expression.KeyConditionBuilder{}, errors.New("cannot use as query")
}

func (a irFieldNe) operandFieldName() string {
	return ""
}

type irFieldBeginsWith struct {
	name   string
	prefix string
}

func (a irFieldBeginsWith) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	if a.name == info.Keys.SortKey {
		return qci.addKey(info, a.name)
	}

	return false
}

func (a irFieldBeginsWith) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	return expression.Name(a.name).BeginsWith(a.prefix), nil
}

func (a irFieldBeginsWith) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	return expression.Key(a.name).BeginsWith(a.prefix), nil
}

func (a irFieldBeginsWith) operandFieldName() string {
	return a.name
}
