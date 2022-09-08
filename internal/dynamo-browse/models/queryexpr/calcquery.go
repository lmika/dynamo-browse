package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
)

func (a *astExpr) calcQuery(info *models.TableInfo) (*models.QueryExecutionPlan, error) {
	root := a.Root

	if root.canBeExecutedAsQuery(info) {
		ke, err := root.calcQueryForQuery(info)
		if err != nil {
			return nil, err
		}

		builder := expression.NewBuilder()
		builder = builder.WithKeyCondition(ke)

		expr, err := builder.Build()
		if err != nil {
			return nil, err
		}

		return &models.QueryExecutionPlan{
			CanQuery:   true,
			Expression: expr,
		}, nil
	}

	cb, err := root.calcQueryForScan(info)
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
