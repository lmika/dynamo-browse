package relselector

type relItemModel struct {
	name string
}

func (ti relItemModel) FilterValue() string {
	return ti.name
}

func (ti relItemModel) Title() string {
	return ti.name
}

func (ti relItemModel) Description() string {
	return ti.name
}
