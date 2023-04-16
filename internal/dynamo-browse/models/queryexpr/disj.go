package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"strings"
)

func (a *astDisjunction) evalToIR(ctx *evalContext, tableInfo *models.TableInfo) (irAtom, error) {
	if len(a.Operands) == 1 {
		return a.Operands[0].evalToIR(ctx, tableInfo)
	}

	conj := make([]irAtom, len(a.Operands))
	for i, op := range a.Operands {
		var err error
		conj[i], err = op.evalToIR(ctx, tableInfo)
		if err != nil {
			return nil, err
		}
	}

	return &irDisjunction{conj: conj}, nil
}

func (a *astDisjunction) evalItem(ctx *evalContext, item models.Item) (exprValue, error) {
	val, err := a.Operands[0].evalItem(ctx, item)
	if err != nil {
		return nil, err
	}
	if len(a.Operands) == 1 {
		return val, nil
	}

	for _, opr := range a.Operands[1:] {
		if isAttributeTrue(val) {
			return boolExprValue(true), nil
		}

		val, err = opr.evalItem(ctx, item)
		if err != nil {
			return nil, err
		}
	}

	return boolExprValue(isAttributeTrue(val)), nil
}

func (a *astDisjunction) canModifyItem(ctx *evalContext, item models.Item) bool {
	if len(a.Operands) == 1 {
		return a.Operands[0].canModifyItem(ctx, item)
	}

	return false
}

func (a *astDisjunction) setEvalItem(ctx *evalContext, item models.Item, value exprValue) error {
	if len(a.Operands) == 1 {
		return a.Operands[0].setEvalItem(ctx, item, value)
	}

	return PathNotSettableError{}
}

func (a *astDisjunction) deleteAttribute(ctx *evalContext, item models.Item) error {
	if len(a.Operands) == 1 {
		return a.Operands[0].deleteAttribute(ctx, item)
	}

	return PathNotSettableError{}
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
	conj []irAtom
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
