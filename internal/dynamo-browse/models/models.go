package models

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

type ResultSet struct {
	Columns []string
	Items   []Item
}

type Item map[string]types.AttributeValue
