package controllers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	bus "github.com/lmika/events"
)

type ColumnsController struct {
	// State
	colModel  *models.Columns
	resultSet *models.ResultSet
}

func NewColumnsController(eventBus *bus.Bus) *ColumnsController {
	cc := &ColumnsController{}

	eventBus.On(newResultSetEvent, cc.onNewResultSet)
	return cc
}

func (cc *ColumnsController) Columns() *models.Columns {
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
	cc.colModel = models.NewColumnsFromResultSet(cc.resultSet)
	return ColumnsUpdated{}
}

func (cc *ColumnsController) onNewResultSet(rs *models.ResultSet) {
	cc.resultSet = rs

	if cc.colModel == nil || !cc.colModel.TableInfo.Equal(rs.TableInfo) {
		cc.colModel = models.NewColumnsFromResultSet(rs)
	}
}
