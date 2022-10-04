package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"strings"
)

func (a *astDisjunction) evalToIR(tableInfo *models.TableInfo) (*irDisjunction, error) {
	conj := make([]*irConjunction, len(a.Operands))
	for i, op := range a.Operands {
		var err error
		conj[i], err = op.evalToIR(tableInfo)
		if err != nil {
			return nil, err
		}
	}

	return &irDisjunction{conj: conj}, nil
}

func (a *astDisjunction) evalItem(item models.Item) (types.AttributeValue, error) {
	if len(a.Operands) == 1 {
		return a.Operands[0].evalItem(item)
	}

	return nil, errors.New("TODO")
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

type irDisjunction struct {
	conj []*irConjunction
}

func (d *irDisjunction) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	// TODO: not entire accurate, as filter expressions are also possible
	if len(d.conj) == 1 {
		return d.conj[0].canBeExecutedAsQuery(info, qci)
	}
	return false
}

func (d *irDisjunction) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	if len(d.conj) == 1 {
		return d.conj[0].calcQueryForQuery(info)
	}

	return expression.KeyConditionBuilder{}, errors.New("expected exactly 1 operand for query")
}

func (d *irDisjunction) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	if len(d.conj) == 1 {
		return d.conj[0].calcQueryForScan(info)
	}

	// TODO: check if can be query
	conds := make([]expression.ConditionBuilder, len(d.conj))
	for i, operand := range d.conj {
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
