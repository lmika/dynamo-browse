package itemrender

type OtherRenderer struct{}

func (u OtherRenderer) TypeName() string {
	return "??"
}

func (sr OtherRenderer) StringValue() string {
	return ""
}

func (u OtherRenderer) MetaInfo() string {
	return "(unrecognised)"
}

func (u OtherRenderer) SubItems() []SubItem {
	return nil
}
