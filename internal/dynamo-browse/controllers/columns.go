package controllers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/dynamo-browse/models"
)

type ColumnsController struct {
	colModel *models.Columns
}

func NewColumnsController() *ColumnsController {
	return &ColumnsController{
		colModel: &models.Columns{
			Columns: []models.Column{
				{Name: "pk", Hidden: false},
				{Name: "sk", Hidden: false},
				{Name: "name", Hidden: true},
				{Name: "address", Hidden: false},
				{Name: "city", Hidden: false},
			},
		},
	}
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
