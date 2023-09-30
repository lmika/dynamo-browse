package models

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"sort"
	"time"
)

type ResultSet struct {
	// Query information
	TableInfo         *TableInfo
	Query             Queryable
	Created           time.Time
	ExclusiveStartKey map[string]types.AttributeValue

	// Result information
	LastEvaluatedKey map[string]types.AttributeValue
	items            []Item
	attributes       []ItemAttribute

	columns []string
}

type Queryable interface {
	String() string
	SerializeToBytes() ([]byte, error)
	HashCode() uint64
	Plan(tableInfo *TableInfo) (*QueryExecutionPlan, error)
}

type ItemAttribute struct {
	Marked bool
	Hidden bool
	Dirty  bool
	New    bool
}

func (rs *ResultSet) NoResults() bool {
	return len(rs.items) == 0 && rs.LastEvaluatedKey == nil
}

func (rs *ResultSet) Items() []Item {
	return rs.items
}

func (rs *ResultSet) SetItems(items []Item) {
	rs.items = items
	rs.attributes = make([]ItemAttribute, len(items))
}

func (rs *ResultSet) AddNewItem(item Item, attrs ItemAttribute) {
	rs.items = append(rs.items, item)
	rs.attributes = append(rs.attributes, attrs)
}

func (rs *ResultSet) SetMark(idx int, marked bool) {
	rs.attributes[idx].Marked = marked
}

func (rs *ResultSet) SetHidden(idx int, hidden bool) {
	rs.attributes[idx].Hidden = hidden
}

func (rs *ResultSet) SetDirty(idx int, dirty bool) {
	rs.attributes[idx].Dirty = dirty
}

func (rs *ResultSet) SetNew(idx int, isNew bool) {
	rs.attributes[idx].New = isNew
}

func (rs *ResultSet) Marked(idx int) bool {
	return rs.attributes[idx].Marked
}

func (rs *ResultSet) Hidden(idx int) bool {
	return rs.attributes[idx].Hidden
}

func (rs *ResultSet) IsDirty(idx int) bool {
	return rs.attributes[idx].Dirty
}

func (rs *ResultSet) IsNew(idx int) bool {
	return rs.attributes[idx].New
}

func (rs *ResultSet) MarkedItems() []ItemIndex {
	items := make([]ItemIndex, 0)
	for i, itemAttr := range rs.attributes {
		if itemAttr.Marked && !itemAttr.Hidden {
			items = append(items, ItemIndex{Index: i, Item: rs.items[i]})
		}
	}
	return items
}

func (rs *ResultSet) Columns() []string {
	if rs.columns == nil {
		rs.RefreshColumns()
	}
	return rs.columns
}

func (rs *ResultSet) RefreshColumns() {
	seenColumns := make(map[string]int)
	seenColumns[rs.TableInfo.Keys.PartitionKey] = 0
	if rs.TableInfo.Keys.SortKey != "" {
		seenColumns[rs.TableInfo.Keys.SortKey] = 1
	}

	for _, definedAttribute := range rs.TableInfo.DefinedAttributes {
		if _, seen := seenColumns[definedAttribute]; !seen {
			seenColumns[definedAttribute] = len(seenColumns)
		}
	}

	otherColsRank := len(seenColumns)
	for _, result := range rs.items {
		for k := range result {
			if _, isSeen := seenColumns[k]; !isSeen {
				seenColumns[k] = otherColsRank
			}
		}
	}

	columns := make([]string, 0, len(seenColumns))
	for k := range seenColumns {
		columns = append(columns, k)
	}
	sort.Slice(columns, func(i, j int) bool {
		if seenColumns[columns[i]] == seenColumns[columns[j]] {
			return columns[i] < columns[j]
		}
		return seenColumns[columns[i]] < seenColumns[columns[j]]
	})

	rs.columns = columns
}

func (rs *ResultSet) HasNextPage() bool {
	return rs.LastEvaluatedKey != nil
}
