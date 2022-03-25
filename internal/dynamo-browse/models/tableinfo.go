package models

type TableInfo struct {
	Name              string
	Keys              KeyAttribute
	DefinedAttributes []string
}

type KeyAttribute struct {
	PartitionKey string
	SortKey      string
}
