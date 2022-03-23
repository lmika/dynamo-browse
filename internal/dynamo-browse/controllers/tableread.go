package controllers

import (
	"context"
	"github.com/lmika/awstools/internal/common/ui/uimodels"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
)

type TableReadController struct {
	tableService *tables.Service
	tableName string
}

func NewTableReadController(tableService *tables.Service, tableName string) *TableReadController {
	return &TableReadController{
		tableService: tableService,
		tableName: tableName,
	}
}

func (c *TableReadController) Scan() uimodels.Operation {
	return uimodels.OperationFn(func(ctx context.Context) error {
		return c.doScan(ctx, false)
	})
}

func (c *TableReadController) doScan(ctx context.Context, quiet bool) error {
	uiCtx := uimodels.Ctx(ctx)

	if !quiet {
		uiCtx.Message("Scanning...")
	}

	resultSet, err := c.tableService.Scan(ctx, c.tableName)
	if err != nil {
		return err
	}

	if !quiet {
		uiCtx.Messagef("Found %d items", len(resultSet.Items))
	}
	uiCtx.Send(NewResultSet{resultSet})
	return nil
}