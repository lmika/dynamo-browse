package models

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

type ResultSet struct {
	Table   string
	Columns []string
	Items   []Item
}

type Item map[string]types.AttributeValue

// Clone creates a clone of the current item
func (i Item) Clone() Item {
	newItem := Item{}

	// TODO: should be a deep clone?
	for k, v := range i {
		newItem[k] = v
	}

	return newItem
}
