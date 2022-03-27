package controllers

import (
	"context"

	"github.com/lmika/awstools/internal/common/ui/uimodels"
	"github.com/lmika/awstools/internal/dynamo-browse/models/modexpr"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/pkg/errors"
)

type TableWriteController struct {
	tableService         *tables.Service
	tableReadControllers *TableReadController
	tableName            string
}

func NewTableWriteController(tableService *tables.Service, tableReadControllers *TableReadController, tableName string) *TableWriteController {
	return &TableWriteController{
		tableService:         tableService,
		tableReadControllers: tableReadControllers,
		tableName:            tableName,
	}
}

func (c *TableWriteController) ToggleReadWrite() uimodels.Operation {
	return uimodels.OperationFn(func(ctx context.Context) error {
		uiCtx := uimodels.Ctx(ctx)
		state := CurrentState(ctx)

		if state.InReadWriteMode {
			uiCtx.Send(SetReadWrite{NewValue: false})
			uiCtx.Message("read/write mode disabled")
		} else {
			uiCtx.Send(SetReadWrite{NewValue: true})
			uiCtx.Message("read/write mode enabled")
		}

		return nil
	})
}

func (c *TableWriteController) Duplicate() uimodels.Operation {
	return uimodels.OperationFn(func(ctx context.Context) error {
		uiCtx := uimodels.Ctx(ctx)
		state := CurrentState(ctx)

		if state.SelectedItem == nil {
			return errors.New("no selected item")
		} else if !state.InReadWriteMode {
			return errors.New("not in read/write mode")
		}

		uiCtx.Input("Dup: ", uimodels.OperationFn(func(ctx context.Context) error {
			modExpr, err := modexpr.Parse(uimodels.PromptValue(ctx))
			if err != nil {
				return err
			}

			newItem, err := modExpr.Patch(state.SelectedItem)
			if err != nil {
				return err
			}

			// TODO: preview new item

			uiCtx := uimodels.Ctx(ctx)
			uiCtx.Input("Put item? ", uimodels.OperationFn(func(ctx context.Context) error {
				if uimodels.PromptValue(ctx) != "y" {
					return errors.New("operation aborted")
				}

				tableInfo, err := c.tableReadControllers.tableInfo(ctx)
				if err != nil {
					return err
				}

				// Delete the item
				if err := c.tableService.Put(ctx, tableInfo, newItem); err != nil {
					return err
				}

				// Rescan to get updated items
				// if err := c.tableReadControllers.doScan(ctx, true); err != nil {
				// 	return err
				// }

				return nil
			}))
			return nil
		}))
		return nil
	})
}

func (c *TableWriteController) Delete() uimodels.Operation {
	return uimodels.OperationFn(func(ctx context.Context) error {
		uiCtx := uimodels.Ctx(ctx)
		state := CurrentState(ctx)

		if state.SelectedItem == nil {
			return errors.New("no selected item")
		} else if !state.InReadWriteMode {
			return errors.New("not in read/write mode")
		}

		uiCtx.Input("Delete item? ", uimodels.OperationFn(func(ctx context.Context) error {
			uiCtx := uimodels.Ctx(ctx)

			if uimodels.PromptValue(ctx) != "y" {
				return errors.New("operation aborted")
			}

			tableInfo, err := c.tableReadControllers.tableInfo(ctx)
			if err != nil {
				return err
			}

			// Delete the item
			if err := c.tableService.Delete(ctx, tableInfo, state.SelectedItem); err != nil {
				return err
			}

			// Rescan to get updated items
			// if err := c.tableReadControllers.doScan(ctx, true); err != nil {
			// 	return err
			// }

			uiCtx.Message("Item deleted")
			return nil
		}))
		return nil
	})
}
