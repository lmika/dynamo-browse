package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
)

func (a *irDisjunction) calcQuery(info *models.TableInfo) (*models.QueryExecutionPlan, error) {
	var qci queryCalcInfo
	if a.canBeExecutedAsQuery(info, &qci) {
		ke, err := a.calcQueryForQuery(info)
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
