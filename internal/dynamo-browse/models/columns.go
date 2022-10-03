package models

import (
	"log"
)

type Columns struct {
	TableInfo *TableInfo
	Columns   []Column
}

func NewColumnsFromResultSet(rs *ResultSet) *Columns {
	cols := make([]Column, len(rs.columns))
	for i, c := range rs.columns {
		cols[i] = Column{Name: c}
	}

	return &Columns{
		TableInfo: rs.TableInfo,
		Columns:   cols,
	}
}

func (cols *Columns) VisibleColumns() []Column {
	if cols == nil {
		return []Column{}
	}

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
