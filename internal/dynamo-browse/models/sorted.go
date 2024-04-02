package models

import (
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/attrutils"
	"sort"
)

// sortedItems is a collection of items that is sorted.
// Items are sorted based on the PK, and SK in ascending order
type sortedItems struct {
	criteria SortCriteria
	items    []Item
}

type SortField struct {
	Field FieldValueEvaluator
	Asc   bool
}

type SortCriteria struct {
	Fields []SortField
}

func (sc SortCriteria) FirstField() SortField {
	if len(sc.Fields) == 0 {
		return SortField{}
	}
	return sc.Fields[0]
}

func (sc SortCriteria) Equals(osc SortCriteria) bool {
	if len(sc.Fields) != len(osc.Fields) {
		return false
	}

	for i := range osc.Fields {
		if sc.Fields[i].Field != osc.Fields[i].Field ||
			sc.Fields[i].Asc != osc.Fields[i].Asc {
			return false
		}
	}

	return true
}

func (sc SortCriteria) Append(osc SortCriteria) SortCriteria {
	newItems := make([]SortField, 0, len(osc.Fields))
	newItems = append(newItems, sc.Fields...)
	newItems = append(newItems, osc.Fields...)
	return SortCriteria{Fields: newItems}
}

func PKSKSortFilter(ti *TableInfo) SortCriteria {
	return SortCriteria{
		Fields: []SortField{
			{Field: SimpleFieldValueEvaluator(ti.Keys.PartitionKey), Asc: true},
			{Field: SimpleFieldValueEvaluator(ti.Keys.SortKey), Asc: true},
		},
	}
}

// Sort sorts the items in place
func Sort(items []Item, criteria SortCriteria) {
	si := sortedItems{items: items, criteria: criteria}
	sort.Sort(&si)
}

func (si *sortedItems) Len() int {
	return len(si.items)
}

func (si *sortedItems) Less(i, j int) bool {
	for _, field := range si.criteria.Fields {
		// Compare primary keys
		pv1, pv2 := field.Field.EvaluateForItem(si.items[i]), field.Field.EvaluateForItem(si.items[j])
		pc, ok := attrutils.CompareScalarAttributes(pv1, pv2)
		if !ok {
			return i < j
		}

		if !field.Asc {
			pc = -pc
		}

		if pc < 0 {
			return true
		} else if pc > 0 {
			return false
		}
	}

	// This should never happen, but just in case
	return i < j
}

func (si *sortedItems) Swap(i, j int) {
	si.items[j], si.items[i] = si.items[i], si.items[j]
}
