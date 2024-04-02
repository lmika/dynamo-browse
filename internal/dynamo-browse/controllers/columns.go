package controllers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/columns"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/evaluators"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/queryexpr"
	bus "github.com/lmika/events"
	"strings"
)

type ColumnsController struct {
	tr *TableReadController

	// State
	colModel  *columns.Columns
	resultSet *models.ResultSet
}

func NewColumnsController(tr *TableReadController, eventBus *bus.Bus) *ColumnsController {
	cc := &ColumnsController{tr: tr}

	eventBus.On(newResultSetEvent, cc.onNewResultSet)
	return cc
}

func (cc *ColumnsController) Columns() *columns.Columns {
	return cc.colModel
}

func (cc *ColumnsController) ToggleVisible(idx int) tea.Msg {
	cc.colModel.Columns[idx].Hidden = !cc.colModel.Columns[idx].Hidden
	return ColumnsUpdated{}
}

func (cc *ColumnsController) ShiftColumnLeft(idx int) tea.Msg {
	if idx == 0 {
		return nil
	}

	col := cc.colModel.Columns[idx-1]
	cc.colModel.Columns[idx-1], cc.colModel.Columns[idx] = cc.colModel.Columns[idx], col
	cc.colModel.WasRearranged = true

	return ColumnsUpdated{}
}

func (cc *ColumnsController) ShiftColumnRight(idx int) tea.Msg {
	if idx >= len(cc.colModel.Columns)-1 {
		return nil
	}

	col := cc.colModel.Columns[idx+1]
	cc.colModel.Columns[idx+1], cc.colModel.Columns[idx] = cc.colModel.Columns[idx], col
	cc.colModel.WasRearranged = true

	return ColumnsUpdated{}
}

func (cc *ColumnsController) SetColumnsToResultSet() tea.Msg {
	cc.colModel = columns.NewColumnsFromResultSet(cc.resultSet)
	return ColumnsUpdated{}
}

func (cc *ColumnsController) onNewResultSet(rs *models.ResultSet, op resultSetUpdateOp) {
	cc.resultSet = rs

	if cc.colModel == nil || (op == resultSetUpdateInit || op == resultSetUpdateQuery) {
		cc.colModel = columns.NewColumnsFromResultSet(rs)
	} else {
		cc.colModel.AddMissingColumns(rs)
	}
}

func (cc *ColumnsController) AddColumn(afterIndex int) tea.Msg {
	return events.PromptForInput("column expr: ", nil, func(value string) tea.Msg {
		colExpr, err := queryexpr.Parse(value)
		if err != nil {
			return events.Error(err)
		}

		newCol := columns.Column{
			Name:      colExpr.String(),
			Evaluator: queryexpr.ExprFieldValueEvaluator{Expr: colExpr},
		}

		if afterIndex >= len(cc.colModel.Columns)-1 {
			cc.colModel.Columns = append(cc.colModel.Columns, newCol)
		} else {
			newCols := make([]columns.Column, 0, len(cc.colModel.Columns)+1)

			newCols = append(newCols, cc.colModel.Columns[:afterIndex+1]...)
			newCols = append(newCols, newCol)
			newCols = append(newCols, cc.colModel.Columns[afterIndex+1:]...)

			cc.colModel.Columns = newCols
		}
		cc.colModel.WasRearranged = true

		return tea.Batch(
			events.SetTeaMessage(ColumnsUpdated{}),
			events.SetTeaMessage(SetSelectedColumnInColSelector(afterIndex+1)),
		)()
	})
}

func (cc *ColumnsController) DeleteColumn(afterIndex int) tea.Msg {
	if len(cc.colModel.Columns) == 0 {
		return nil
	}

	newCols := make([]columns.Column, 0, len(cc.colModel.Columns)-1)
	newCols = append(newCols, cc.colModel.Columns[:afterIndex]...)
	newCols = append(newCols, cc.colModel.Columns[afterIndex+1:]...)
	cc.colModel.Columns = newCols
	cc.colModel.WasRearranged = true

	return ColumnsUpdated{}
}

func (cc *ColumnsController) SortByColumn(index int) tea.Msg {
	if index >= len(cc.colModel.Columns) {
		return nil
	}

	column := cc.colModel.Columns[index]
	newCriteria := models.SortCriteria{
		Fields: []models.SortField{
			{Field: column.Evaluator, Asc: true},
		},
	}
	if ff := cc.SortCriteria().FirstField(); evaluators.Equals(ff.Field, column.Evaluator) {
		newCriteria.Fields[0].Asc = !ff.Asc
	}

	cc.SetSortCriteria(newCriteria)
	return ColumnsUpdated{}
}

func (c *ColumnsController) AttributesWithPrefix(prefix string) []string {
	options := make([]string, 0)
	for _, col := range c.resultSet.Columns() {
		if strings.HasPrefix(col, prefix) {
			options = append(options, col)
		}
	}
	return options
}

func (cc *ColumnsController) SortCriteria() models.SortCriteria {
	if cc.resultSet == nil {
		return models.SortCriteria{}
	}

	return cc.resultSet.SortCriteria()
}

func (cc *ColumnsController) SetSortCriteria(criteria models.SortCriteria) {
	cc.tr.SortResultSet(criteria)
}
