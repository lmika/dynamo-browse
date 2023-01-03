package controllers

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/queryexpr"
	"github.com/lmika/audax/internal/dynamo-browse/models/serialisable"
	"github.com/lmika/audax/internal/dynamo-browse/services/itemrenderer"
	"github.com/lmika/audax/internal/dynamo-browse/services/viewsnapshot"
	bus "github.com/lmika/events"
	"github.com/pkg/errors"
	"golang.design/x/clipboard"
	"log"
	"strings"
	"sync"
)

type resultSetUpdateOp int

const (
	resultSetUpdateInit resultSetUpdateOp = iota
	resultSetUpdateQuery
	resultSetUpdateFilter
	resultSetUpdateSnapshotRestore
	resultSetUpdateRescan
	resultSetUpdateTouch
	resultSetUpdateScript
)

type MarkOp int

const (
	MarkOpMark MarkOp = iota
	MarkOpUnmark
	MarkOpToggle
)

type TableReadController struct {
	tableService        TableReadService
	workspaceService    *viewsnapshot.ViewSnapshotService
	itemRendererService *itemrenderer.Service
	jobController       *JobsController
	eventBus            *bus.Bus
	tableName           string
	loadFromLastView    bool

	// state
	mutex         *sync.Mutex
	state         *State
	clipboardInit bool
}

func NewTableReadController(
	state *State,
	tableService TableReadService,
	workspaceService *viewsnapshot.ViewSnapshotService,
	itemRendererService *itemrenderer.Service,
	jobController *JobsController,
	eventBus *bus.Bus,
	tableName string,
) *TableReadController {
	return &TableReadController{
		state:               state,
		tableService:        tableService,
		workspaceService:    workspaceService,
		itemRendererService: itemRendererService,
		jobController:       jobController,
		eventBus:            eventBus,
		tableName:           tableName,
		mutex:               new(sync.Mutex),
	}
}

// Init does an initial scan of the table.  If no table is specified, it prompts for a table, then does a scan.
func (c *TableReadController) Init() tea.Msg {
	// Restore previous view
	if c.loadFromLastView {
		if vs, err := c.workspaceService.ViewRestore(); err == nil && vs != nil {
			return c.updateViewToSnapshot(vs)
		}
	}

	if c.tableName == "" {
		return c.ListTables(true)
	} else {
		return c.ScanTable(c.tableName)
	}
}

func (c *TableReadController) ListTables(quitIfNoTable bool) tea.Msg {
	return NewJob(c.jobController, "Listing tables…", func(ctx context.Context) (any, error) {
		tables, err := c.tableService.ListTables(context.Background())
		if err != nil {
			return nil, err
		}
		return tables, nil
	}).OnDone(func(res any) tea.Msg {
		return PromptForTableMsg{
			Tables: res.([]string),
			OnSelected: func(tableName string) tea.Msg {
				if tableName == "" {
					if quitIfNoTable {
						return tea.Quit()
					}
					return events.StatusMsg("No table selected")
				}

				return c.ScanTable(tableName)
			},
		}
	}).Submit()
}

func (c *TableReadController) ScanTable(name string) tea.Msg {
	return NewJob(c.jobController, "Scanning…", func(ctx context.Context) (*models.ResultSet, error) {
		tableInfo, err := c.tableService.Describe(ctx, name)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot describe %v", c.tableName)
		}

		resultSet, err := c.tableService.Scan(ctx, tableInfo)
		if resultSet != nil {
			resultSet = c.tableService.Filter(resultSet, c.state.Filter())
		}

		return resultSet, err
	}).OnEither(c.handleResultSetFromJobResult(c.state.Filter(), true, resultSetUpdateInit)).Submit()
}

func (c *TableReadController) PromptForQuery() tea.Msg {
	return events.PromptForInputMsg{
		Prompt: "query: ",
		OnDone: func(value string) tea.Msg {
			resultSet := c.state.ResultSet()
			if resultSet == nil {
				return events.StatusMsg("Result-set is nil")
			}

			return c.runQuery(resultSet.TableInfo, value, "", true)
		},
	}
}

func (c *TableReadController) runQuery(tableInfo *models.TableInfo, query, newFilter string, pushSnapshot bool) tea.Msg {
	if query == "" {
		return NewJob(c.jobController, "Scanning…", func(ctx context.Context) (*models.ResultSet, error) {
			newResultSet, err := c.tableService.ScanOrQuery(context.Background(), tableInfo, nil)

			if newResultSet != nil && newFilter != "" {
				newResultSet = c.tableService.Filter(newResultSet, newFilter)
			}

			return newResultSet, err
		}).OnEither(c.handleResultSetFromJobResult(newFilter, pushSnapshot, resultSetUpdateQuery)).Submit()
	}

	expr, err := queryexpr.Parse(query)
	if err != nil {
		return events.Error(err)
	}

	return c.doIfNoneDirty(func() tea.Msg {
		return NewJob(c.jobController, "Running query…", func(ctx context.Context) (*models.ResultSet, error) {
			newResultSet, err := c.tableService.ScanOrQuery(context.Background(), tableInfo, expr)

			if newFilter != "" && newResultSet != nil {
				newResultSet = c.tableService.Filter(newResultSet, newFilter)
			}
			return newResultSet, err
		}).OnEither(c.handleResultSetFromJobResult(newFilter, pushSnapshot, resultSetUpdateQuery)).Submit()
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
		OnDone: func(value string) tea.Msg {
			if value != "y" {
				return events.StatusMsg("operation aborted")
			}

			return cmd()
		},
	}
}

func (c *TableReadController) Rescan() tea.Msg {
	return c.doIfNoneDirty(func() tea.Msg {
		resultSet := c.state.ResultSet()
		return c.doScan(resultSet, resultSet.Query, true, resultSetUpdateRescan)
	})
}

func (c *TableReadController) doScan(resultSet *models.ResultSet, query models.Queryable, pushBackstack bool, op resultSetUpdateOp) tea.Msg {
	return NewJob(c.jobController, "Rescan…", func(ctx context.Context) (*models.ResultSet, error) {
		newResultSet, err := c.tableService.ScanOrQuery(ctx, resultSet.TableInfo, query)
		if newResultSet != nil {
			newResultSet = c.tableService.Filter(newResultSet, c.state.Filter())
		}

		return newResultSet, err
	}).OnEither(c.handleResultSetFromJobResult(c.state.Filter(), pushBackstack, op)).Submit()
}

func (c *TableReadController) setResultSetAndFilter(resultSet *models.ResultSet, filter string, pushBackstack bool, op resultSetUpdateOp) tea.Msg {
	if resultSet != nil && pushBackstack {
		details := serialisable.ViewSnapshotDetails{
			TableName: resultSet.TableInfo.Name,
			Filter:    filter,
		}
		if q := resultSet.Query; q != nil {
			details.Query = q.String()
		}

		if err := c.workspaceService.PushSnapshot(details); err != nil {
			log.Printf("cannot push snapshot: %v", err)
		}
	}

	c.state.setResultSetAndFilter(resultSet, filter)

	c.eventBus.Fire(newResultSetEvent, resultSet, op)

	return c.state.buildNewResultSetMessage("")
}

func (c *TableReadController) Mark(op MarkOp) tea.Msg {
	c.state.withResultSet(func(resultSet *models.ResultSet) {
		for i := range resultSet.Items() {
			if resultSet.Hidden(i) {
				continue
			}

			switch op {
			case MarkOpMark:
				resultSet.SetMark(i, true)
			case MarkOpUnmark:
				resultSet.SetMark(i, false)
			case MarkOpToggle:
				resultSet.SetMark(i, !resultSet.Marked(i))
			}
		}
	})
	return ResultSetUpdated{}
}

func (c *TableReadController) Filter() tea.Msg {
	return events.PromptForInputMsg{
		Prompt: "filter: ",
		OnDone: func(value string) tea.Msg {
			resultSet := c.state.ResultSet()
			if resultSet == nil {
				return events.StatusMsg("Result-set is nil")
			}

			return NewJob(c.jobController, "Applying Filter…", func(ctx context.Context) (*models.ResultSet, error) {
				newResultSet := c.tableService.Filter(resultSet, value)
				return newResultSet, nil
			}).OnEither(c.handleResultSetFromJobResult(value, true, resultSetUpdateFilter)).Submit()
		},
	}
}

func (c *TableReadController) handleResultSetFromJobResult(filter string, pushbackStack bool, op resultSetUpdateOp) func(newResultSet *models.ResultSet, err error) tea.Msg {
	return func(newResultSet *models.ResultSet, err error) tea.Msg {
		if err == nil {
			return c.setResultSetAndFilter(newResultSet, filter, pushbackStack, op)
		}

		var partialResultsErr models.PartialResultsError
		if errors.As(err, &partialResultsErr) {
			if newResultSet == nil {
				return events.StatusMsg("Operation cancelled")
			}

			return events.Confirm(applyToN("View the ", len(newResultSet.Items()), "item", "items", " returned so far? "), func(yes bool) tea.Msg {
				if yes {
					return c.setResultSetAndFilter(newResultSet, filter, pushbackStack, op)
				}
				return events.StatusMsg("Operation cancelled")
			})
		}

		if newResultSet != nil {
			return c.setResultSetAndFilter(newResultSet, filter, pushbackStack, op)
		}
		return events.Error(err)
	}
}

func (c *TableReadController) ViewBack() tea.Msg {
	viewSnapshot, err := c.workspaceService.ViewBack()
	if err != nil {
		return events.Error(err)
	} else if viewSnapshot == nil {
		return events.StatusMsg("Backstack is empty")
	}

	return c.updateViewToSnapshot(viewSnapshot)
}

func (c *TableReadController) ViewForward() tea.Msg {
	viewSnapshot, err := c.workspaceService.ViewForward()
	if err != nil {
		return events.Error(err)
	} else if viewSnapshot == nil {
		return events.StatusMsg("At top of view stack")
	}

	return c.updateViewToSnapshot(viewSnapshot)
}

func (c *TableReadController) updateViewToSnapshot(viewSnapshot *serialisable.ViewSnapshot) tea.Msg {
	var err error
	currentResultSet := c.state.ResultSet()

	if currentResultSet == nil {
		return NewJob(c.jobController, "Fetching table info…", func(ctx context.Context) (*models.TableInfo, error) {
			tableInfo, err := c.tableService.Describe(context.Background(), viewSnapshot.Details.TableName)
			if err != nil {
				return nil, err
			}
			return tableInfo, nil
		}).OnDone(func(tableInfo *models.TableInfo) tea.Msg {
			return c.runQuery(tableInfo, viewSnapshot.Details.Query, viewSnapshot.Details.Filter, false)
		}).Submit()
	}

	var currentQueryExpr string
	if currentResultSet.Query != nil {
		currentQueryExpr = currentResultSet.Query.String()
	}

	if viewSnapshot.Details.TableName == currentResultSet.TableInfo.Name && viewSnapshot.Details.Query == currentQueryExpr {
		return NewJob(c.jobController, "Applying filter…", func(ctx context.Context) (*models.ResultSet, error) {
			return c.tableService.Filter(currentResultSet, viewSnapshot.Details.Filter), nil
		}).OnEither(c.handleResultSetFromJobResult(viewSnapshot.Details.Filter, false, resultSetUpdateSnapshotRestore)).Submit()
	}

	return NewJob(c.jobController, "Running query…", func(ctx context.Context) (tea.Msg, error) {
		tableInfo := currentResultSet.TableInfo
		if viewSnapshot.Details.TableName != currentResultSet.TableInfo.Name {
			tableInfo, err = c.tableService.Describe(context.Background(), viewSnapshot.Details.TableName)
			if err != nil {
				return nil, err
			}
		}

		return c.runQuery(tableInfo, viewSnapshot.Details.Query, viewSnapshot.Details.Filter, false), nil
	}).OnDone(func(m tea.Msg) tea.Msg {
		return m
	}).Submit()
}

func (c *TableReadController) CopyItemToClipboard(idx int) tea.Msg {
	if err := c.initClipboard(); err != nil {
		return events.Error(err)
	}

	itemCount := 0
	c.state.withResultSet(func(resultSet *models.ResultSet) {
		sb := new(strings.Builder)
		_ = applyToMarkedItems(resultSet, idx, func(idx int, item models.Item) error {
			if sb.Len() > 0 {
				fmt.Fprintln(sb, "---")
			}
			c.itemRendererService.RenderItem(sb, resultSet.Items()[idx], resultSet, true)
			itemCount += 1
			return nil
		})
		clipboard.Write(clipboard.FmtText, []byte(sb.String()))
	})

	return events.StatusMsg(applyToN("", itemCount, "item", "items", " copied to clipboard"))
}

func (c *TableReadController) initClipboard() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.clipboardInit {
		return nil
	}

	if err := clipboard.Init(); err != nil {
		return errors.Wrap(err, "unable to enable clipboard")
	}
	c.clipboardInit = true
	return nil
}
