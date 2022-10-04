package tableselect

import "github.com/charmbracelet/bubbles/list"

type tableItem struct {
	name string
}

func (ti tableItem) FilterValue() string {
	return ti.name
}

func (ti tableItem) Title() string {
	return ti.name
}

func (ti tableItem) Description() string {
	return ""
}

func toListItems(xs []string) []list.Item {
	ls := make([]list.Item, len(xs))
	for i, x := range xs {
		ls[i] = tableItem{name: x}
	}
	return ls
}
