package controllers

import (
	"context"
	"encoding/csv"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"os"
	"sync"
)

type TableReadController struct {
	tableService TableReadService
	tableName    string

	// state
	mutex *sync.Mutex
	state *State
	//resultSet *models.ResultSet
	//filter    string
}

func NewTableReadController(state *State, tableService TableReadService, tableName string) *TableReadController {
	return &TableReadController{
		state:        state,
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

		return c.setResultSetAndFilter(resultSet, c.state.Filter())
	}
}

func (c *TableReadController) PromptForQuery() tea.Cmd {
	return func() tea.Msg {
		return events.PromptForInputMsg{
			Prompt: "query: ",
			OnDone: func(value string) tea.Cmd {
				if value == "" {
					return c.Rescan()
				}

				return func() tea.Msg {
					resultSet := c.state.ResultSet()
					newResultSet, err := c.tableService.ScanOrQuery(context.Background(), resultSet.TableInfo, value)
					if err != nil {
						return events.Error(err)
					}

					return c.setResultSetAndFilter(newResultSet, "")
				}
			},
		}
	}
}

func (c *TableReadController) Rescan() tea.Cmd {
	return func() tea.Msg {
		return c.doScan(context.Background(), c.state.ResultSet())
	}
}

func (c *TableReadController) ExportCSV(filename string) tea.Cmd {
	return func() tea.Msg {
		resultSet := c.state.ResultSet()
		if resultSet == nil {
			return events.Error(errors.New("no result set"))
		}

		f, err := os.Create(filename)
		if err != nil {
			return events.Error(errors.Wrapf(err, "cannot export to '%v'", filename))
		}
		defer f.Close()

		cw := csv.NewWriter(f)
		defer cw.Flush()

		columns := resultSet.Columns
		if err := cw.Write(columns); err != nil {
			return events.Error(errors.Wrapf(err, "cannot export to '%v'", filename))
		}

		row := make([]string, len(resultSet.Columns))
		for _, item := range resultSet.Items() {
			for i, col := range columns {
				row[i], _ = item.AttributeValueAsString(col)
			}
			if err := cw.Write(row); err != nil {
				return events.Error(errors.Wrapf(err, "cannot export to '%v'", filename))
			}
		}

		return nil
	}
}

func (c *TableReadController) doScan(ctx context.Context, resultSet *models.ResultSet) tea.Msg {
	newResultSet, err := c.tableService.Scan(ctx, resultSet.TableInfo)
	if err != nil {
		return events.Error(err)
	}

	newResultSet = c.tableService.Filter(newResultSet, c.state.Filter())

	return c.setResultSetAndFilter(newResultSet, c.state.Filter())
}

func (c *TableReadController) setResultSetAndFilter(resultSet *models.ResultSet, filter string) tea.Msg {
	c.state.setResultSetAndFilter(resultSet, filter)

	var statusMessage string
	if filter != "" {
		var filteredCount int
		for i := range resultSet.Items() {
			if !resultSet.Hidden(i) {
				filteredCount += 1
			}
		}
		statusMessage = fmt.Sprintf("%d of %d items returned", filteredCount, len(resultSet.Items()))
	} else {
		statusMessage = fmt.Sprintf("%d items returned", len(resultSet.Items()))
	}

	return NewResultSet{resultSet, statusMessage}
}

func (c *TableReadController) Unmark() tea.Cmd {
	return func() tea.Msg {
		c.state.withResultSet(func(resultSet *models.ResultSet) {
			for i := range resultSet.Items() {
				resultSet.SetMark(i, false)
			}
		})
		return ResultSetUpdated{}
	}
}

func (c *TableReadController) Filter() tea.Cmd {
	return func() tea.Msg {
		return events.PromptForInputMsg{
			Prompt: "filter: ",
			OnDone: func(value string) tea.Cmd {
				return func() tea.Msg {
					resultSet := c.state.ResultSet()
					newResultSet := c.tableService.Filter(resultSet, value)

					return c.setResultSetAndFilter(newResultSet, value)
				}
			},
		}
	}
}
