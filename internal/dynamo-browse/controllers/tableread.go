package controllers

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"sync"
)

type TableReadController struct {
	tableService TableReadService
	tableName    string

	// state
	mutex     *sync.Mutex
	resultSet *models.ResultSet
	filter    string
}

func NewTableReadController(tableService TableReadService, tableName string) *TableReadController {
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

		return c.setResultSetAndFilter(resultSet, c.filter)
	}
}

func (c *TableReadController) Rescan() tea.Cmd {
	return func() tea.Msg {
		return c.doScan(context.Background(), c.resultSet)
	}
}

func (c *TableReadController) doScan(ctx context.Context, resultSet *models.ResultSet) tea.Msg {
	newResultSet, err := c.tableService.Scan(ctx, resultSet.TableInfo)
	if err != nil {
		return events.Error(err)
	}

	newResultSet = c.tableService.Filter(newResultSet, c.filter)

	return c.setResultSetAndFilter(newResultSet, c.filter)
}

func (c *TableReadController) ResultSet() *models.ResultSet {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.resultSet
}

func (c *TableReadController) setResultSetAndFilter(resultSet *models.ResultSet, filter string) tea.Msg {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.resultSet = resultSet
	c.filter = filter
	return NewResultSet{resultSet}
}

func (c *TableReadController) Unmark() tea.Cmd {
	return func() tea.Msg {
		resultSet := c.ResultSet()

		for i := range resultSet.Items() {
			resultSet.SetMark(i, false)
		}

		c.mutex.Lock()
		defer c.mutex.Unlock()

		c.resultSet = resultSet
		return ResultSetUpdated{}
	}
}

func (c *TableReadController) Filter() tea.Cmd {
	return func() tea.Msg {
		return events.PromptForInputMsg{
			Prompt: "filter: ",
			OnDone: func(value string) tea.Cmd {
				return func() tea.Msg {
					resultSet := c.ResultSet()
					newResultSet := c.tableService.Filter(resultSet, value)

					return c.setResultSetAndFilter(newResultSet, value)
				}
			},
		}
	}
}
