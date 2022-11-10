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

	switch a.Op {
	case "<":
		return irFieldCmp{name: nameIR, value: valueIR, cmpType: cmpTypeLt}, nil
	case "<=":
		return irFieldCmp{name: nameIR, value: valueIR, cmpType: cmpTypeLe}, nil
	case ">":
		return irFieldCmp{name: nameIR, value: valueIR, cmpType: cmpTypeGt}, nil
	case ">=":
		return irFieldCmp{name: nameIR, value: valueIR, cmpType: cmpTypeGe}, nil
	}

	return nil, errors.Errorf("unrecognised operator: %v", a.Op)
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

type irFieldCmp struct {
	name    irNamePath
	value   irValue
	cmpType int
}

func (a irFieldCmp) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	keyName := a.name.keyName()
	if keyName == "" {
		return false
	}

	if keyName == info.Keys.SortKey {
		return qci.addKey(info, keyName)
	}

	return false
}

func (a irFieldCmp) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
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

func (a irFieldCmp) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
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

//func (a irFieldCmp) operandFieldName() string {
//	return a.name
//}
