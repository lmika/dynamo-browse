package itemrender

type OtherRenderer struct{}

func (u OtherRenderer) TypeName() string {
	return "(other)"
}

func (u OtherRenderer) StringValue() string {
	return "(other)"
}

func (u OtherRenderer) SubItems() []SubItem {
	return nil
}
