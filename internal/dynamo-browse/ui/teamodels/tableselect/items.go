package tableselect

import "github.com/charmbracelet/bubbles/list"

type tableItem struct {
	name string
}

func (ti tableItem) FilterValue() string {
	return ""
}

func (ti tableItem) Title() string {
	return ti.name
}

func (ti tableItem) Description() string {
	return "abc"
}

func toListItems[T list.Item](xs []T) []list.Item {
	ls := make([]list.Item, len(xs))
	for i, x := range xs {
		ls[i] = x
	}
	return ls
}
