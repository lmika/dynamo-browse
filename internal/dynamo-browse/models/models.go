package models

type ResultSet struct {
	TableInfo  *TableInfo
	Columns    []string
	items      []Item
	attributes []ItemAttribute
}

type ItemAttribute struct {
	Marked bool
	Hidden bool
}

func (rs *ResultSet) Items() []Item {
	return rs.items
}

func (rs *ResultSet) SetItems(items []Item) {
	rs.items = items
	rs.attributes = make([]ItemAttribute, len(items))
}

func (rs *ResultSet) SetMark(idx int, marked bool) {
	rs.attributes[idx].Marked = marked
}

func (rs *ResultSet) SetHidden(idx int, hidden bool) {
	rs.attributes[idx].Hidden = hidden
}

func (rs *ResultSet) Marked(idx int) bool {
	return rs.attributes[idx].Marked
}

func (rs *ResultSet) Hidden(idx int) bool {
	return rs.attributes[idx].Hidden
}

func (rs *ResultSet) MarkedItems() []Item {
	items := make([]Item, 0)
	for i, itemAttr := range rs.attributes {
		if itemAttr.Marked && !itemAttr.Hidden {
			items = append(items, rs.items[i])
		}
	}
	return items
}
