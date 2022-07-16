package models

type ItemType string

const (
	UnsetItemType  ItemType = ""
	StringItemType ItemType = "S"
	NumberItemType ItemType = "N"
	BoolItemType   ItemType = "BOOL"
	NullItemType   ItemType = "NULL"
)
