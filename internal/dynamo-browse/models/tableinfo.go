package models

type TableInfo struct {
	Name              string
	Keys              KeyAttribute
	DefinedAttributes []string
	GSIs              []TableGSI
}

type TableGSI struct {
	Name string
	Keys KeyAttribute
}

func (ti *TableInfo) Equal(other *TableInfo) bool {
	return ti.Name == other.Name &&
		ti.Keys.PartitionKey == other.Keys.PartitionKey &&
		ti.Keys.SortKey == other.Keys.SortKey &&
		len(ti.DefinedAttributes) == len(other.DefinedAttributes) // Probably should be all
}

type KeyAttribute struct {
	PartitionKey string
	SortKey      string
}
