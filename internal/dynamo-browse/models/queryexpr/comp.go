package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

func (a *astComparisonOp) evalToIR(info *models.TableInfo) (irAtom, error) {
	if a.Op == "" {
		return a.Ref.evalToIR(info)
	}

	v, err := a.Value.rightOperandGoValue()
	if err != nil {
		return nil, err
	}

	singleName, isSingleName := a.Ref.leftOperandName()
	if !isSingleName {
		return nil, errors.Errorf("%v: cannot use dereferences", singleName)
	}

	switch a.Op {
	case "<":
		return irFieldCmp{name: singleName, value: v, cmpType: cmpTypeLt}, nil
	case "<=":
		return irFieldCmp{name: singleName, value: v, cmpType: cmpTypeLe}, nil
	case ">":
		return irFieldCmp{name: singleName, value: v, cmpType: cmpTypeGt}, nil
	case ">=":
		return irFieldCmp{name: singleName, value: v, cmpType: cmpTypeGe}, nil
	}

	return nil, errors.Errorf("unrecognised operator: %v", a.Op)
}

func (a *astComparisonOp) leftOperandName() (string, bool) {
	return a.Ref.leftOperandName()
}

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
	name    string
	value   any
	cmpType int
}

func (a irFieldCmp) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	if a.name == info.Keys.SortKey {
		return qci.addKey(info, a.name)
	}

	return false
}

func (a irFieldCmp) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	switch a.cmpType {
	case cmpTypeLt:
		return expression.Name(a.name).LessThan(expression.Value(a.value)), nil
	case cmpTypeLe:
		return expression.Name(a.name).LessThanEqual(expression.Value(a.value)), nil
	case cmpTypeGt:
		return expression.Name(a.name).GreaterThan(expression.Value(a.value)), nil
	case cmpTypeGe:
		return expression.Name(a.name).GreaterThanEqual(expression.Value(a.value)), nil
	}
	return expression.ConditionBuilder{}, errors.New("unsupported cmp type")
}

func (a irFieldCmp) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	switch a.cmpType {
	case cmpTypeLt:
		return expression.Key(a.name).LessThan(expression.Value(a.value)), nil
	case cmpTypeLe:
		return expression.Key(a.name).LessThanEqual(expression.Value(a.value)), nil
	case cmpTypeGt:
		return expression.Key(a.name).GreaterThan(expression.Value(a.value)), nil
	case cmpTypeGe:
		return expression.Key(a.name).GreaterThanEqual(expression.Value(a.value)), nil
	}
	return expression.KeyConditionBuilder{}, errors.New("unsupported cmp type")
}

func (a irFieldCmp) operandFieldName() string {
	return a.name
}
