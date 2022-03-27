package controllers

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
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

// Init does an initial scan of the table.  If no table is specified, it prompts for a table, then does a scan.
func (c *TableReadController) Init() tea.Cmd {
	if c.tableName == "" {
		return c.listTables()
	} else {
		return c.scanTable(c.tableName)
	}
}

func (c *TableReadController) listTables() tea.Cmd {
	return func() tea.Msg {
		tables, err := c.tableService.ListTables(context.Background())
		if err != nil {
			return events.Error(err)
		}

		return PromptForTableMsg{
			Tables: tables,
			OnSelected: func(tableName string) tea.Cmd {
				return c.scanTable(tableName)
			},
		}
	}
}

func (c *TableReadController) scanTable(name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		log.Println("Fetching table info")
		tableInfo, err := c.tableService.Describe(ctx, name)
		if err != nil {
			return events.Error(errors.Wrapf(err, "cannot describe %v", c.tableName))
		}

		log.Println("Scanning")
		resultSet, err := c.tableService.Scan(ctx, tableInfo)
		if err != nil {
			log.Println("error: ", err)
			return events.Error(err)
		}

		log.Println("Scan done")
		return NewResultSet{resultSet}
	}
}

func (c *TableReadController) Rescan(resultSet *models.ResultSet) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		log.Println("Scanning")
		resultSet, err := c.tableService.Scan(ctx, resultSet.TableInfo)
		if err != nil {
			log.Println("error: ", err)
			return events.Error(err)
		}

		log.Println("Scan done")
		return NewResultSet{resultSet}
	}
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
