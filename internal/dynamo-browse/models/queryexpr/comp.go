package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/attrutils"
	"github.com/pkg/errors"
)

func (a *astComparisonOp) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	leftIR, err := a.Ref.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}

	if a.Op == "" {
		return leftIR, nil
	}

	cmpType, hasCmpType := opToCmdType[a.Op]
	if !hasCmpType {
		return nil, errors.Errorf("unrecognised operator: %v", a.Op)
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

	nameIR, isNameIR := leftIR.(nameIRAtom)
	valueIR, isValueIR := rightIR.(valueIRAtom)
	if isNameIR && isValueIR {
		return irKeyFieldCmp{nameIR, valueIR, cmpType}, nil
	}

	return irGenericCmp{leftOpr, rightOpr, cmpType}, nil
}

func (a *astComparisonOp) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
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

	cmp, isComparable := attrutils.CompareScalarAttributes(left, right)
	if !isComparable {
		return nil, ValuesNotComparable{Left: left, Right: right}
	}

	switch opToCmdType[a.Op] {
	case cmpTypeLt:
		return &types.AttributeValueMemberBOOL{Value: cmp < 0}, nil
	case cmpTypeLe:
		return &types.AttributeValueMemberBOOL{Value: cmp <= 0}, nil
	case cmpTypeGt:
		return &types.AttributeValueMemberBOOL{Value: cmp > 0}, nil
	case cmpTypeGe:
		return &types.AttributeValueMemberBOOL{Value: cmp >= 0}, nil
	}
	return nil, errors.Errorf("unrecognised operator: %v", a.Op)
}

func (a *astComparisonOp) canModifyItem(item models.Item) bool {
	if a.Op != "" {
		return false
	}
	return a.Ref.canModifyItem(item)
}

func (a *astComparisonOp) setEvalItem(item models.Item, value types.AttributeValue) error {
	if a.Op != "" {
		return PathNotSettableError{}
	}
	return a.Ref.setEvalItem(item, value)
}

func (a *astComparisonOp) deleteAttribute(item models.Item) error {
	if a.Op != "" {
		return PathNotSettableError{}
	}
	return a.Ref.deleteAttribute(item)

}

func (a *astComparisonOp) String() string {
	if a.Op == "" {
		return a.Ref.String()
	}
	return a.Ref.String() + a.Op + a.Value.String()
}

const (
	cmpTypeLt int = 0
	cmpTypeLe int = 1
	cmpTypeGt int = 2
	cmpTypeGe int = 3
)

var opToCmdType = map[string]int{
	"<":  cmpTypeLt,
	"<=": cmpTypeLe,
	">":  cmpTypeGt,
	">=": cmpTypeGe,
}

type irKeyFieldCmp struct {
	name    nameIRAtom
	value   valueIRAtom
	cmpType int
}

func (a irKeyFieldCmp) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	keyName := a.name.keyName()
	if keyName == "" {
		return false
	}

	if keyName == info.Keys.SortKey {
		return qci.addKey(info, keyName)
	}

	return false
}

func (a irKeyFieldCmp) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	nb := a.name.calcName(info)
	vb := a.value.goValue()

	switch a.cmpType {
	case cmpTypeLt:
		return nb.LessThan(expression.Value(vb)), nil
	case cmpTypeLe:
		return nb.LessThanEqual(expression.Value(vb)), nil
	case cmpTypeGt:
		return nb.GreaterThan(expression.Value(vb)), nil
	case cmpTypeGe:
		return nb.GreaterThanEqual(expression.Value(vb)), nil
	}
	return expression.ConditionBuilder{}, errors.New("unsupported cmp type")
}

func (a irKeyFieldCmp) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	keyName := a.name.keyName()
	vb := a.value.goValue()

	switch a.cmpType {
	case cmpTypeLt:
		return expression.Key(keyName).LessThan(expression.Value(vb)), nil
	case cmpTypeLe:
		return expression.Key(keyName).LessThanEqual(expression.Value(vb)), nil
	case cmpTypeGt:
		return expression.Key(keyName).GreaterThan(expression.Value(vb)), nil
	case cmpTypeGe:
		return expression.Key(keyName).GreaterThanEqual(expression.Value(vb)), nil
	}
	return expression.KeyConditionBuilder{}, errors.New("unsupported cmp type")
}

type irGenericCmp struct {
	left    oprIRAtom
	right   oprIRAtom
	cmpType int
}

func (a irGenericCmp) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	nb := a.left.calcOperand(info)
	vb := a.right.calcOperand(info)

	switch a.cmpType {
	case cmpTypeLt:
		return expression.LessThan(nb, vb), nil
	case cmpTypeLe:
		return expression.LessThanEqual(nb, vb), nil
	case cmpTypeGt:
		return expression.GreaterThan(nb, vb), nil
	case cmpTypeGe:
		return expression.GreaterThanEqual(nb, vb), nil
	}
	return expression.ConditionBuilder{}, errors.New("unsupported cmp type")
}
