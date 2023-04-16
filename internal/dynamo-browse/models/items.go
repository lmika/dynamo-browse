package models

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/attrutils"
)

type ItemIndex struct {
	Index int
	Item  Item
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

func (i Item) KeyValue(info *TableInfo) map[string]types.AttributeValue {
	itemKey := make(map[string]types.AttributeValue)
	itemKey[info.Keys.PartitionKey] = i[info.Keys.PartitionKey]
	if info.Keys.SortKey != "" {
		itemKey[info.Keys.SortKey] = i[info.Keys.SortKey]
	}
	return itemKey
}

func (i Item) AttributeValueAsString(key string) (string, bool) {
	return attrutils.AttributeToString(i[key])
}
