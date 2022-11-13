package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

func (a *astComparisonOp) evalToIR(info *models.TableInfo) (irAtom, error) {
	leftIR, err := a.Ref.evalToIR(info)
	if err != nil {
		return nil, err
	}

	if a.Op == "" {
		return leftIR, nil
	}

	cmpType, hasCmpType := opToCmdType[a.Op]
	if !hasCmpType {
		errors.Errorf("unrecognised operator: %v", a.Op)
	}

	leftOpr, isLeftOpr := leftIR.(oprIRAtom)
	if !isLeftOpr {
		return nil, OperandNotAnOperandError{}
	}

	rightIR, err := a.Value.evalToIR(info)
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

	//if a.Op == "" {
	//	return a.Ref.evalToIR(info)
	//}
	//
	//v, err := a.Value.rightOperandGoValue()
	//if err != nil {
	//	return nil, err
	//}
	//
	//singleName, isSingleName := a.Ref.leftOperandName()
	//if !isSingleName {
	//	return nil, errors.Errorf("%v: cannot use dereferences", singleName)
	//}

	return irGenericCmp{leftOpr, rightOpr, cmpType}, nil
}

//func (a *astComparisonOp) leftOperandName() (string, bool) {
//	return a.Ref.leftOperandName()
//}

func (a *astComparisonOp) String() string {
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

func (a irGenericCmp) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	return false
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

func (a irGenericCmp) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	return expression.KeyConditionBuilder{}, errors.New("unsupported cmp type")
}
