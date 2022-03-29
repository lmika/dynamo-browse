package controllers

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/pkg/errors"
	"sync"
)

type TableReadController struct {
	tableService *tables.Service
	tableName    string

	// state
	mutex     *sync.Mutex
	resultSet *models.ResultSet
}

func NewTableReadController(tableService *tables.Service, tableName string) *TableReadController {
	return &TableReadController{
		tableService: tableService,
		tableName:    tableName,
		mutex:        new(sync.Mutex),
	}
}

// Init does an initial scan of the table.  If no table is specified, it prompts for a table, then does a scan.
func (c *TableReadController) Init() tea.Cmd {
	if c.tableName == "" {
		return c.ListTables()
	} else {
		return c.ScanTable(c.tableName)
	}
}

func (c *TableReadController) ListTables() tea.Cmd {
	return func() tea.Msg {
		tables, err := c.tableService.ListTables(context.Background())
		if err != nil {
			return events.Error(err)
		}

		return PromptForTableMsg{
			Tables: tables,
			OnSelected: func(tableName string) tea.Cmd {
				return c.ScanTable(tableName)
			},
		}
	}
}

func (c *TableReadController) ScanTable(name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		tableInfo, err := c.tableService.Describe(ctx, name)
		if err != nil {
			return events.Error(errors.Wrapf(err, "cannot describe %v", c.tableName))
		}

		resultSet, err := c.tableService.Scan(ctx, tableInfo)
		if err != nil {
			return events.Error(err)
		}

		return c.setResultSet(resultSet)
	}
}

func (c *TableReadController) Rescan() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		resultSet, err := c.tableService.Scan(ctx, c.resultSet.TableInfo)
		if err != nil {
			return events.Error(err)
		}

		return c.setResultSet(resultSet)
	}
}

func (c *TableReadController) setResultSet(resultSet *models.ResultSet) tea.Msg {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.resultSet = resultSet
	return NewResultSet{resultSet}
}

/*
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
*/

// tableInfo returns the table info from the state if a result set exists.  If not, it fetches the
// table information from the service.
// func (c *TableReadController) tableInfo(ctx context.Context) (*models.TableInfo, error) {
// 	/*
// 		if existingResultSet := CurrentState(ctx).ResultSet; existingResultSet != nil {
// 			return existingResultSet.TableInfo, nil
// 		}
// 	*/

// 	tableInfo, err := c.tableService.Describe(ctx, c.tableName)
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "cannot describe %v", c.tableName)
// 	}
// 	return tableInfo, nil
// }
