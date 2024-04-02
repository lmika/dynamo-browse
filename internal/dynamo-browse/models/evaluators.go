package models

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type FieldValueEvaluator interface {
	EvaluateForItem(item Item) types.AttributeValue
}

type SimpleFieldValueEvaluator string

func (sfve SimpleFieldValueEvaluator) EvaluateForItem(item Item) types.AttributeValue {
	return item[string(sfve)]
}
