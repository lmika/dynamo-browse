package models

import (
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/attrutils"
	"sort"
)

// sortedItems is a collection of items that is sorted.
// Items are sorted based on the PK, and SK in ascending order
type sortedItems struct {
	tableInfo *TableInfo
	items     []Item
}

// Sort sorts the items in place
func Sort(items []Item, tableInfo *TableInfo) {
	si := sortedItems{items: items, tableInfo: tableInfo}
	sort.Sort(&si)
}

func (si *sortedItems) Len() int {
	return len(si.items)
}

func (si *sortedItems) Less(i, j int) bool {
	// Compare primary keys
	pv1, pv2 := si.items[i][si.tableInfo.Keys.PartitionKey], si.items[j][si.tableInfo.Keys.PartitionKey]
	pc, ok := attrutils.CompareScalarAttributes(pv1, pv2)
	if !ok {
		return i < j
	}

	if pc < 0 {
		return true
	} else if pc > 0 {
		return false
	}

	// Partition keys are equal, compare sort key
	if sortKey := si.tableInfo.Keys.SortKey; sortKey != "" {
		sv1, sv2 := si.items[i][sortKey], si.items[j][sortKey]
		sc, ok := attrutils.CompareScalarAttributes(sv1, sv2)
		if !ok {
			return i < j
		}

		if sc < 0 {
			return true
		} else if sc > 0 {
			return false
		}
	}

	// This should never happen, but just in case
	return i < j
}

func (si *sortedItems) Swap(i, j int) {
	si.items[j], si.items[i] = si.items[i], si.items[j]
}
