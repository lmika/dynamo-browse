package models

import "log"

type Columns struct {
	TableInfo *TableInfo
	Columns   []Column
}

func (cols *Columns) VisibleColumns() []Column {
	visibleCols := make([]Column, 0)
	for _, col := range cols.Columns {
		if col.Hidden {
			continue
		}
		visibleCols = append(visibleCols, col)
	}
	log.Printf("%v --> %v", cols.Columns, visibleCols)
	return visibleCols
}

type Column struct {
	Name   string
	Hidden bool
}
