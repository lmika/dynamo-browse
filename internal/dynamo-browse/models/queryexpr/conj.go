package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"math/big"
	"strings"
)

func (a *astConjunction) evalToIR(ctx *evalContext, tableInfo *models.TableInfo) (irAtom, error) {
	if len(a.Operands) == 1 {
		return a.Operands[0].evalToIR(ctx, tableInfo)
	} else if len(a.Operands) == 2 {
		left, err := a.Operands[0].evalToIR(ctx, tableInfo)
		if err != nil {
			return nil, err
		}

		right, err := a.Operands[1].evalToIR(ctx, tableInfo)
		if err != nil {
			return nil, err
		}

		return &irDualConjunction{left: left, right: right}, nil
	}

	atoms := make([]irAtom, len(a.Operands))
	for i, op := range a.Operands {
		var err error
		atoms[i], err = op.evalToIR(ctx, tableInfo)
		if err != nil {
			return nil, err
		}
	}

	return &irMultiConjunction{atoms: atoms}, nil
}

func (a *astConjunction) evalItem(ctx *evalContext, item models.Item) (exprValue, error) {
	val, err := a.Operands[0].evalItem(ctx, item)
	if err != nil {
		return nil, err
	}
	if len(a.Operands) == 1 {
		return val, nil
	}

	for _, opr := range a.Operands[1:] {
		if !isAttributeTrue(val) {
			return boolExprValue(false), nil
		}

		val, err = opr.evalItem(ctx, item)
		if err != nil {
			return nil, err
		}
	}

	return boolExprValue(isAttributeTrue(val)), nil
}

func (a *astConjunction) canModifyItem(ctx *evalContext, item models.Item) bool {
	if len(a.Operands) == 1 {
		return a.Operands[0].canModifyItem(ctx, item)
	}

	return false
}

func (a *astConjunction) setEvalItem(ctx *evalContext, item models.Item, value exprValue) error {
	if len(a.Operands) == 1 {
		return a.Operands[0].setEvalItem(ctx, item, value)
	}

	return PathNotSettableError{}
}

func (a *astConjunction) deleteAttribute(ctx *evalContext, item models.Item) error {
	if len(a.Operands) == 1 {
		return a.Operands[0].deleteAttribute(ctx, item)
	}

	return PathNotSettableError{}
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

type irDualConjunction struct {
	left     irAtom
	right    irAtom
	leftIsPK bool
}

func (i *irDualConjunction) canBeExecutedAsQuery(qci *queryCalcInfo) bool {
	qciCopy := qci.clone()

	leftCanExecuteAsQuery := canExecuteAsQuery(i.left, qci)
	if leftCanExecuteAsQuery {
		i.leftIsPK = qci.hasSeenPrimaryKey()
		return canExecuteAsQuery(i.right, qci)
	}

	// Might be that the right is the partition key, so test again with them swapped
	rightCanExecuteAsQuery := canExecuteAsQuery(i.right, qciCopy)
	if rightCanExecuteAsQuery {
		return canExecuteAsQuery(i.left, qciCopy)
	}

	return false
}

func (i *irDualConjunction) calcQueryForQuery() (expression.KeyConditionBuilder, error) {
	left, err := i.left.(queryableIRAtom).calcQueryForQuery()
	if err != nil {
		return expression.KeyConditionBuilder{}, err
	}

	right, err := i.right.(queryableIRAtom).calcQueryForQuery()
	if err != nil {
		return expression.KeyConditionBuilder{}, err
	}

	if i.leftIsPK {
		return expression.KeyAnd(left, right), nil
	}
	return expression.KeyAnd(right, left), nil
}

func (i *irDualConjunction) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	left, err := i.left.calcQueryForScan(info)
	if err != nil {
		return expression.ConditionBuilder{}, err
	}

	right, err := i.right.calcQueryForScan(info)
	if err != nil {
		return expression.ConditionBuilder{}, err
	}

	return expression.And(left, right), nil
}

type irMultiConjunction struct {
	atoms []irAtom
}

func (d *irMultiConjunction) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	conds := make([]expression.ConditionBuilder, len(d.atoms))
	for i, operand := range d.atoms {
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

func isAttributeTrue(attr exprValue) bool {
	switch val := attr.(type) {
	case nullExprValue:
		return false
	case boolExprValue:
		return bool(val)
	case stringableExprValue:
		return val.asString() != ""
	case numberableExprValue:
		return val.asBigFloat().Cmp(&big.Float{}) != 0
	}
	return true
}
