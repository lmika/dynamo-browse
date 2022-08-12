package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

func (a *astExpr) calcQuery(tableInfo *models.TableInfo) (*models.QueryExecutionPlan, error) {
	return a.Root.calcQuery(tableInfo)
}

func (d *astDisjunction) calcQuery(info *models.TableInfo) (*models.QueryExecutionPlan, error) {
	// TODO: check if can be a query
	cb, err := d.calcQueryForScan(info)
	if err != nil {
		return nil, err
	}

	builder := expression.NewBuilder()
	builder = builder.WithFilter(cb)

	expr, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return &models.QueryExecutionPlan{
		CanQuery:   false,
		Expression: expr,
	}, nil
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

func (a *astBinOp) calcQuery(info *models.TableInfo) (*models.QueryExecutionPlan, error) {
	// TODO: check if can be a query
	cb, err := a.calcQueryForScan(info)
	if err != nil {
		return nil, err
	}

	builder := expression.NewBuilder()
	builder = builder.WithFilter(cb)

	expr, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return &models.QueryExecutionPlan{
		CanQuery:   false,
		Expression: expr,
	}, nil
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
