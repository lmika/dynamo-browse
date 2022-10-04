package controllers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/columns"
	"github.com/lmika/audax/internal/dynamo-browse/models/queryexpr"
	bus "github.com/lmika/events"
)

type ColumnsController struct {
	// State
	colModel  *columns.Columns
	resultSet *models.ResultSet
}

func NewColumnsController(eventBus *bus.Bus) *ColumnsController {
	cc := &ColumnsController{}

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

	return ColumnsUpdated{}
}

func (cc *ColumnsController) ShiftColumnRight(idx int) tea.Msg {
	if idx >= len(cc.colModel.Columns)-1 {
		return nil
	}

	col := cc.colModel.Columns[idx+1]
	cc.colModel.Columns[idx+1], cc.colModel.Columns[idx] = cc.colModel.Columns[idx], col

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
	}
}

func (cc *ColumnsController) AddColumn(afterIndex int) tea.Msg {
	return events.PromptForInput("column expr: ", func(value string) tea.Msg {
		colExpr, err := queryexpr.Parse(value)
		if err != nil {
			return events.Error(err)
		}

		newCol := columns.Column{
			Name:      colExpr.String(),
			Evaluator: columns.ExprFieldValueEvaluator{Expr: colExpr},
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

		return ColumnsUpdated{}
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

	return ColumnsUpdated{}
}
