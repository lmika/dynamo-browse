package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

func (a *astBinOp) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	// If this is the partition key, then the op must be equals
	if a.Name == info.Keys.PartitionKey && a.Op == "=" {
		return qci.addKey(info, a.Name)
	}

	// If this is sort key, then the op must be equals (and others)
	if a.Name == info.Keys.SortKey && (a.Op == "=" || a.Op == "^=") {
		return qci.addKey(info, a.Name)
	}

	return false
}

func (a *astBinOp) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	v, err := a.Value.goValue()
	if err != nil {
		return expression.ConditionBuilder{}, err
	}

	switch a.Op {
	case "=":
		return expression.Name(a.Name).Equal(expression.Value(v)), nil
	case "^=":
		strValue, isStrValue := v.(string)
		if !isStrValue {
			return expression.ConditionBuilder{}, errors.New("operand '^=' must be string")
		}
		return expression.Name(a.Name).BeginsWith(strValue), nil
	}

	return expression.ConditionBuilder{}, errors.Errorf("unrecognised operator: %v", a.Op)
}

func (a *astBinOp) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	v, err := a.Value.goValue()
	if err != nil {
		return expression.KeyConditionBuilder{}, err
	}

	switch a.Op {
	case "=":
		return expression.Key(a.Name).Equal(expression.Value(v)), nil
	case "^=":
		strValue, isStrValue := v.(string)
		if !isStrValue {
			return expression.KeyConditionBuilder{}, errors.New("operand '^=' must be string")
		}
		return expression.Key(a.Name).BeginsWith(strValue), nil
	}

	return expression.KeyConditionBuilder{}, errors.Errorf("unrecognised operator: %v", a.Op)
}

func (a *astBinOp) queryKeyName() string {
	return a.Name
}

func (a *astBinOp) String() string {
	return a.Name + a.Op + a.Value.String()
}
