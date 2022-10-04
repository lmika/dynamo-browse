package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

func (a *astBinOp) evalToIR(info *models.TableInfo) (irAtom, error) {
	v, err := a.Value.goValue()
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
	case "^=":
		strValue, isStrValue := v.(string)
		if !isStrValue {
			return nil, errors.New("operand '^=' must be string")
		}
		return irFieldBeginsWith{name: singleName, prefix: strValue}, nil
	}

	return nil, errors.Errorf("unrecognised operator: %v", a.Op)
}

func (a *astBinOp) evalItem(item models.Item) (types.AttributeValue, error) {
	left, err := a.Ref.evalItem(item)
	if err != nil {
		return nil, err
	}

	if a.Op == "" {
		return left, nil
	}

	panic("TODO")
}

func (a *astBinOp) String() string {
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
