package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"strings"
)

func (d *astConjunction) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	switch len(d.Operands) {
	case 1:
		return d.Operands[0].queryKeyName() == info.Keys.PartitionKey && d.Operands[0].canBeExecutedAsQuery(info, qci)
	case 2:
		return d.Operands[0].canBeExecutedAsQuery(info, qci) && d.Operands[1].canBeExecutedAsQuery(info, qci)
	}
	return false
}

func (d *astConjunction) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	if len(d.Operands) == 1 {
		return d.Operands[0].calcQueryForQuery(info)
	} else if len(d.Operands) != 2 {
		return expression.KeyConditionBuilder{}, errors.Errorf("internal error: expected len to be either 1 or 2, but was %v", len(d.Operands))
	}

	left, err := d.Operands[0].calcQueryForQuery(info)
	if err != nil {
		return expression.KeyConditionBuilder{}, err
	}

	right, err := d.Operands[1].calcQueryForQuery(info)
	if err != nil {
		return expression.KeyConditionBuilder{}, err
	}

	if d.Operands[0].queryKeyName() == info.Keys.PartitionKey {
		return expression.KeyAnd(left, right), nil
	}
	return expression.KeyAnd(right, left), nil
}

func (d *astConjunction) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	if len(d.Operands) == 1 {
		return d.Operands[0].calcQueryForScan(info)
	}

	// TODO: check if can be query
	conds := make([]expression.ConditionBuilder, len(d.Operands))
	for i, operand := range d.Operands {
		cond, err := operand.calcQueryForScan(info)
		if err != nil {
			return expression.ConditionBuilder{}, err
		}
		conds[i] = cond
	}

	// Build conjunction
	conjExpr := expression.And(conds[0], conds[1], conds[2:]...)
	return conjExpr, nil
}

func (d *astConjunction) String() string {
	sb := new(strings.Builder)
	for i, operand := range d.Operands {
		if i > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(operand.String())
	}
	return sb.String()
}
