package models

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

type ResultSet struct {
	TableInfo *TableInfo
	Columns   []string
	Items     []Item
	Marks     map[int]bool
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

func (rs *ResultSet) SetMark(idx int, marked bool) {
	if marked {
		if rs.Marks == nil {
			rs.Marks = make(map[int]bool)
		}
		rs.Marks[idx] = true
	} else {
		delete(rs.Marks, idx)
	}
}

func (rs *ResultSet) Marked(idx int) bool {
	return rs.Marks[idx]
}

func (rs *ResultSet) MarkedItems() []Item {
	items := make([]Item, 0)
	for i, marked := range rs.Marks {
		if marked {
			items = append(items, rs.Items[i])
		}
	}
	return items
}
