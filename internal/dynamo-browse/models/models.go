package models

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

type ResultSet struct {
	Table   string
	Columns []string
	Items   []Item
}

type Item map[string]types.AttributeValue
