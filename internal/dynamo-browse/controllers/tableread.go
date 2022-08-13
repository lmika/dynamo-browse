package controllers

import (
	"context"
	"encoding/csv"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/queryexpr"
	"github.com/lmika/audax/internal/dynamo-browse/services/workspaces"
	"github.com/pkg/errors"
	"log"
	"os"
	"sync"
)

type TableReadController struct {
	tableService     TableReadService
	workspaceService *workspaces.ViewSnapshotService
	tableName        string

	// state
	mutex *sync.Mutex
	state *State
}

func NewTableReadController(state *State, tableService TableReadService, workspaceService *workspaces.ViewSnapshotService, tableName string) *TableReadController {
	return &TableReadController{
		state:            state,
		tableService:     tableService,
		workspaceService: workspaceService,
		tableName:        tableName,
		mutex:            new(sync.Mutex),
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
		resultSet = c.tableService.Filter(resultSet, c.state.Filter())

		return c.setResultSetAndFilter(resultSet, c.state.Filter(), true)
	}
}

func (c *TableReadController) PromptForQuery() tea.Cmd {
	return func() tea.Msg {
		return events.PromptForInputMsg{
			Prompt: "query: ",
			OnDone: func(value string) tea.Cmd {
				return func() tea.Msg {
					return c.runQuery(c.state.ResultSet().TableInfo, value, "", true)
				}
			},
		}
	}
}

func (c *TableReadController) runQuery(tableInfo *models.TableInfo, query, newFilter string, pushSnapshot bool) tea.Msg {
	if query == "" {
		newResultSet, err := c.tableService.ScanOrQuery(context.Background(), tableInfo, nil)
		if err != nil {
			return events.Error(err)
		}

		if newFilter != "" {
			newResultSet = c.tableService.Filter(newResultSet, newFilter)
		}

		return c.setResultSetAndFilter(newResultSet, newFilter, pushSnapshot)
	}

	expr, err := queryexpr.Parse(query)
	if err != nil {
		return events.SetError(err)
	}

	return c.doIfNoneDirty(func() tea.Msg {
		newResultSet, err := c.tableService.ScanOrQuery(context.Background(), tableInfo, expr)
		if err != nil {
			return events.Error(err)
		}

		if newFilter != "" {
			newResultSet = c.tableService.Filter(newResultSet, newFilter)
		}
		return c.setResultSetAndFilter(newResultSet, newFilter, pushSnapshot)
	})
}

func (c *TableReadController) doIfNoneDirty(cmd tea.Cmd) tea.Msg {
	var anyDirty = false
	for i := 0; i < len(c.state.ResultSet().Items()); i++ {
		anyDirty = anyDirty || c.state.ResultSet().IsDirty(i)
	}

	if !anyDirty {
		return cmd()
	}

	return events.PromptForInputMsg{
		Prompt: "reset modified items? ",
		OnDone: func(value string) tea.Cmd {
			if value != "y" {
				return events.SetStatus("operation aborted")
			}

			return cmd
		},
	}
}

func (c *TableReadController) Rescan() tea.Cmd {
	return func() tea.Msg {
		return c.doIfNoneDirty(func() tea.Msg {
			resultSet := c.state.ResultSet()
			return c.doScan(context.Background(), resultSet, resultSet.Query, true)
		})
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

		columns := resultSet.Columns()
		if err := cw.Write(columns); err != nil {
			return events.Error(errors.Wrapf(err, "cannot export to '%v'", filename))
		}

		row := make([]string, len(columns))
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

func (c *TableReadController) doScan(ctx context.Context, resultSet *models.ResultSet, query models.Queryable, pushBackstack bool) tea.Msg {
	newResultSet, err := c.tableService.ScanOrQuery(ctx, resultSet.TableInfo, query)
	if err != nil {
		return events.Error(err)
	}

	newResultSet = c.tableService.Filter(newResultSet, c.state.Filter())

	return c.setResultSetAndFilter(newResultSet, c.state.Filter(), pushBackstack)
}

func (c *TableReadController) setResultSetAndFilter(resultSet *models.ResultSet, filter string, pushBackstack bool) tea.Msg {
	if pushBackstack {
		if err := c.workspaceService.PushSnapshot(resultSet, filter); err != nil {
			log.Printf("cannot push snapshot: %v", err)
		}
	}

	c.state.setResultSetAndFilter(resultSet, filter)
	return c.state.buildNewResultSetMessage("")
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

					return c.setResultSetAndFilter(newResultSet, value, true)
				}
			},
		}
	}
}

func (c *TableReadController) ViewBack() tea.Cmd {
	return func() tea.Msg {
		viewSnapshot, err := c.workspaceService.PopSnapshot()
		if err != nil {
			return events.Error(err)
		} else if viewSnapshot == nil {
			return events.StatusMsg("Backstack is empty")
		}

		currentResultSet := c.state.ResultSet()

		var currentQueryExpr string
		if currentResultSet.Query != nil {
			currentQueryExpr = currentResultSet.Query.String()
		}

		if viewSnapshot.TableName == currentResultSet.TableInfo.Name && viewSnapshot.Query == currentQueryExpr {
			log.Printf("backstack: setting filter to '%v'", viewSnapshot.Filter)

			newResultSet := c.tableService.Filter(currentResultSet, viewSnapshot.Filter)
			return c.setResultSetAndFilter(newResultSet, viewSnapshot.Filter, false)
		}

		tableInfo := currentResultSet.TableInfo
		if viewSnapshot.TableName != currentResultSet.TableInfo.Name {
			tableInfo, err = c.tableService.Describe(context.Background(), viewSnapshot.TableName)
			if err != nil {
				return events.Error(err)
			}
		}

		log.Printf("backstack: running query: table = '%v', query = '%v', filter = '%v'",
			tableInfo.Name, viewSnapshot.Query, viewSnapshot.Filter)
		return c.runQuery(tableInfo, viewSnapshot.Query, viewSnapshot.Filter, false)
	}
}
