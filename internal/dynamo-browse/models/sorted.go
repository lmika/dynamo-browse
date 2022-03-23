package models

import "sort"

// sortedItems is a collection of items that is sorted.
// Items are sorted based on the PK, and SK in ascending order
type sortedItems struct {
	pk, sk string
	items  []Item
}

// Sort sorts the items in place
func Sort(items []Item, pk, sk string) {
	si := sortedItems{items: items, pk: pk, sk: sk}
	sort.Sort(&si)
}

func (si *sortedItems) Len() int {
	return len(si.items)
}

func (si *sortedItems) Less(i, j int) bool {
	// Compare primary keys
	pv1, pv2 := si.items[i][si.pk], si.items[j][si.pk]
	pc, ok := compareScalarAttributes(pv1, pv2)
	if !ok {
		return i < j
	}

	if pc < 0 {
		return true
	} else if pc > 0 {
		return false
	}

	// Partition keys are equal, compare sort key
	sv1, sv2 := si.items[i][si.sk], si.items[j][si.sk]
	sc, ok := compareScalarAttributes(sv1, sv2)
	if !ok {
		return i < j
	}

	if sc < 0 {
		return true
	} else if sc > 0 {
		return false
	}

	// This should never happen, but just in case
	return i < j
}

func (si *sortedItems) Swap(i, j int) {
	si.items[j], si.items[i] = si.items[i], si.items[j]
}
