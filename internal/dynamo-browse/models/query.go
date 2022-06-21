package models

import "github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"

type QueryExecutionPlan struct {
	CanQuery   bool
	Expression expression.Expression
}
