package controllers

import (
	"github.com/lmika/audax/internal/dynamo-browse/models"
)

type ColumnsController struct {
	colModel *models.Columns
}

func NewColumnsController() *ColumnsController {
	return &ColumnsController{
		colModel: &models.Columns{
			Columns: []string{"pk", "sk", "city"},
		},
	}
}

func (cc *ColumnsController) Columns() *models.Columns {
	return cc.colModel
}
