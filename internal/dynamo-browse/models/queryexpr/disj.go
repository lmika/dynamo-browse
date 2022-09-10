package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"strings"
)

func (d *astDisjunction) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	// TODO: not entire accurate, as filter expressions are also possible
	if len(d.Operands) == 1 {
		return d.Operands[0].canBeExecutedAsQuery(info, qci)
	}
	return false
}

func (d *astDisjunction) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	if len(d.Operands) == 1 {
		return d.Operands[0].calcQueryForQuery(info)
	}

	return expression.KeyConditionBuilder{}, errors.New("expected exactly 1 operand for query")
}

func (d *astDisjunction) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
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

	// Build disjunction
	disjExpr := expression.Or(conds[0], conds[1], conds[2:]...)
	return disjExpr, nil
}

func (d *astDisjunction) String() string {
	sb := new(strings.Builder)
	for i, operand := range d.Operands {
		if i > 0 {
			sb.WriteString(" or ")
		}
		sb.WriteString(operand.String())
	}
	return sb.String()
}
