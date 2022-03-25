package controllers

import (
	"context"

	"github.com/lmika/awstools/internal/common/ui/uimodels"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/pkg/errors"
)

type TableReadController struct {
	tableService *tables.Service
	tableName    string
}

func NewTableReadController(tableService *tables.Service, tableName string) *TableReadController {
	return &TableReadController{
		tableService: tableService,
		tableName:    tableName,
	}
}

func (c *TableReadController) Scan() uimodels.Operation {
	return uimodels.OperationFn(func(ctx context.Context) error {
		return c.doScan(ctx, false)
	})
}

func (c *TableReadController) doScan(ctx context.Context, quiet bool) (err error) {
	uiCtx := uimodels.Ctx(ctx)

	if !quiet {
		uiCtx.Message("Scanning...")
	}

	tableInfo, err := c.tableInfo(ctx)
	if err != nil {
		return err
	}

	resultSet, err := c.tableService.Scan(ctx, tableInfo)
	if err != nil {
		return err
	}

	if !quiet {
		uiCtx.Messagef("Found %d items", len(resultSet.Items))
	}
	uiCtx.Send(NewResultSet{resultSet})
	return nil
}

// tableInfo returns the table info from the state if a result set exists.  If not, it fetches the
// table information from the service.
func (c *TableReadController) tableInfo(ctx context.Context) (*models.TableInfo, error) {
	if existingResultSet := CurrentState(ctx).ResultSet; existingResultSet != nil {
		return existingResultSet.TableInfo, nil
	}

	tableInfo, err := c.tableService.Describe(ctx, c.tableName)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot describe %v", c.tableName)
	}
	return tableInfo, nil
}
