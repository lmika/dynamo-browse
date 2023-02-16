package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/attrutils"
	"github.com/pkg/errors"
	"strings"
)

func (a *astEqualityOp) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	leftIR, err := a.Ref.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}

	if a.Op == "" {
		return leftIR, nil
	}

	leftOpr, isLeftOpr := leftIR.(oprIRAtom)
	if !isLeftOpr {
		return nil, OperandNotAnOperandError{}
	}

	rightIR, err := a.Value.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}

	rightOpr, isRightIR := rightIR.(oprIRAtom)
	if !isRightIR {
		return nil, OperandNotAnOperandError{}
	}

	switch a.Op {
	case "=":
		nameIR, isNameIR := leftIR.(nameIRAtom)
		valueIR, isValueIR := rightIR.(valueIRAtom)
		if isNameIR && isValueIR {
			return irKeyFieldEq{name: nameIR, value: valueIR}, nil
		}
		return irGenericEq{name: leftOpr, value: rightOpr}, nil
	case "!=":
		return irFieldNe{name: leftOpr, value: rightOpr}, nil
	case "^=":
		nameIR, isNameIR := leftIR.(nameIRAtom)
		if !isNameIR {
			return nil, OperandNotANameError(a.Ref.String())
		}
		realValueIR, isRealValueIR := rightIR.(irValue)
		if !isRealValueIR {
			return nil, ValueMustBeLiteralError{}
		}
		return irFieldBeginsWith{name: nameIR, value: realValueIR}, nil
	}

	return nil, errors.Errorf("unrecognised operator: %v", a.Op)
}

func (a *astEqualityOp) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
	left, err := a.Ref.evalItem(ctx, item)
	if err != nil {
		return nil, err
	}

	if a.Op == "" {
		return left, nil
	}

	right, err := a.Value.evalItem(ctx, item)
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
	case "!=":
		cmp, isComparable := attrutils.CompareScalarAttributes(left, right)
		if !isComparable {
			return nil, ValuesNotComparable{Left: left, Right: right}
		}
		return &types.AttributeValueMemberBOOL{Value: cmp != 0}, nil
	case "^=":
		strValue, isStrValue := right.(*types.AttributeValueMemberS)
		if !isStrValue {
			return nil, errors.New("operand '^=' must be string")
		}

		leftAsStr, canBeString := attrutils.AttributeToString(left)
		if !canBeString {
			return nil, ValueNotConvertableToString{Val: left}
		}
		return &types.AttributeValueMemberBOOL{Value: strings.HasPrefix(leftAsStr, strValue.Value)}, nil
	}

	return nil, errors.Errorf("unrecognised operator: %v", a.Op)
}

func (a *astEqualityOp) canModifyItem(ctx *evalContext, item models.Item) bool {
	if a.Op != "" {
		return false
	}
	return a.Ref.canModifyItem(ctx, item)
}

func (a *astEqualityOp) setEvalItem(ctx *evalContext, item models.Item, value types.AttributeValue) error {
	if a.Op != "" {
		return PathNotSettableError{}
	}
	return a.Ref.setEvalItem(ctx, item, value)
}

func (a *astEqualityOp) deleteAttribute(ctx *evalContext, item models.Item) error {
	if a.Op != "" {
		return PathNotSettableError{}
	}
	return a.Ref.deleteAttribute(ctx, item)
}

func (a *astEqualityOp) String() string {
	if a.Op == "" {
		return a.Ref.String()
	}
	return a.Ref.String() + a.Op + a.Value.String()
}

type irKeyFieldEq struct {
	name  nameIRAtom
	value valueIRAtom
}

func (a irKeyFieldEq) canBeExecutedAsQuery(qci *queryCalcInfo) bool {
	keyName := a.name.keyName()
	if keyName == "" {
		return false
	}

	if keyName == qci.keysUnderTest.PartitionKey ||
		(keyName == qci.keysUnderTest.SortKey && qci.hasSeenPrimaryKey()) {
		return qci.addKey(keyName)
	}

	return false
}

func (a irKeyFieldEq) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	nb := a.name.calcName(info)
	vb := a.value.calcOperand(info)
	return nb.Equal(vb), nil
}

func (a irKeyFieldEq) calcQueryForQuery() (expression.KeyConditionBuilder, error) {
	vb := a.value.goValue()
	return expression.Key(a.name.keyName()).Equal(expression.Value(vb)), nil
}

type irGenericEq struct {
	name  oprIRAtom
	value oprIRAtom
}

func (a irGenericEq) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	nb := a.name.calcOperand(info)
	vb := a.value.calcOperand(info)
	return expression.Equal(nb, vb), nil
}

type irFieldNe struct {
	name  oprIRAtom
	value oprIRAtom
}

func (a irFieldNe) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	nb := a.name.calcOperand(info)
	vb := a.value.calcOperand(info)
	return expression.NotEqual(nb, vb), nil
}

type irFieldBeginsWith struct {
	name  nameIRAtom
	value irValue
}

func (a irFieldBeginsWith) canBeExecutedAsQuery(qci *queryCalcInfo) bool {
	keyName := a.name.keyName()
	if keyName == "" {
		return false
	}

	if keyName == qci.keysUnderTest.SortKey && qci.hasSeenPrimaryKey() {
		return qci.addKey(a.name.keyName())
	}

	return false
}

func (a irFieldBeginsWith) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	nb := a.name.calcName(info)
	vb := a.value.goValue()
	strValue, isStrValue := vb.(string)
	if !isStrValue {
		return expression.ConditionBuilder{}, errors.New("operand '^=' must be string")
	}

	return nb.BeginsWith(strValue), nil
}

func (a irFieldBeginsWith) calcQueryForQuery() (expression.KeyConditionBuilder, error) {
	vb := a.value.goValue()
	strValue, isStrValue := vb.(string)
	if !isStrValue {
		return expression.KeyConditionBuilder{}, errors.New("operand '^=' must be string")
	}

	return expression.Key(a.name.keyName()).BeginsWith(strValue), nil
}
