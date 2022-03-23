package controllers

import (
	"context"
	"github.com/lmika/awstools/internal/common/ui/uimodels"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

type TableWriteController struct {
	tableService *tables.Service
	tableReadControllers *TableReadController
	tableName string
}

func NewTableWriteController(tableService *tables.Service, tableReadControllers *TableReadController, tableName string) *TableWriteController {
	return &TableWriteController{
		tableService: tableService,
		tableReadControllers: tableReadControllers,
		tableName: tableName,
	}
}

func (c *TableWriteController) Delete(item models.Item) uimodels.Operation {
	return uimodels.OperationFn(func(ctx context.Context) error {
		uiCtx := uimodels.Ctx(ctx)

		// TODO: only do if rw mode enabled

		uiCtx.Input("Delete item?", uimodels.OperationFn(func(ctx context.Context) error {
			uiCtx := uimodels.Ctx(ctx)

			if uimodels.PromptValue(ctx) != "y" {
				return errors.New("operation aborted")
			}

			// Delete the item
			if err := c.tableService.Delete(ctx, c.tableName, item); err != nil {
				return err
			}

			// Rescan to get updated items
			if err := c.tableReadControllers.doScan(ctx, true); err != nil {
				return err
			}

			uiCtx.Message("Item deleted")
			return nil
		}))
		return nil
	})
}
