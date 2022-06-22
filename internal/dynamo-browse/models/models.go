package models

type ResultSet struct {
	TableInfo  *TableInfo
	Query      Queryable
	Columns    []string
	items      []Item
	attributes []ItemAttribute
}

type Queryable interface {
	String() string
	Plan(tableInfo *TableInfo) (*QueryExecutionPlan, error)
}

type ItemAttribute struct {
	Marked bool
	Hidden bool
	Dirty  bool
	New    bool
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

func (rs *ResultSet) MarkedItems() []Item {
	items := make([]Item, 0)
	for i, itemAttr := range rs.attributes {
		if itemAttr.Marked && !itemAttr.Hidden {
			items = append(items, rs.items[i])
		}
	}
	return items
}
