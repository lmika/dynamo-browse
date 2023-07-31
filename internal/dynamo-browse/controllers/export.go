package controllers

import (
	"context"
	"encoding/csv"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/attrutils"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/jobs"
	"github.com/pkg/errors"
	"os"
)

type ExportController struct {
	state         *State
	tableService  TableReadService
	jobController *JobsController
	columns       *ColumnsController
}

func NewExportController(state *State, tableService TableReadService, jobsController *JobsController, columns *ColumnsController) *ExportController {
	return &ExportController{state, tableService, jobsController, columns}
}

func (c *ExportController) ExportCSV(filename string, opts ExportOptions) tea.Msg {
	resultSet := c.state.ResultSet()
	if resultSet == nil {
		return events.Error(errors.New("no result set"))
	}

	return NewJob(c.jobController, fmt.Sprintf("Exporting to %v…", filename), func(ctx context.Context) (int, error) {
		f, err := os.Create(filename)
		if err != nil {
			return 0, errors.Wrapf(err, "cannot export to '%v'", filename)
		}
		defer f.Close()

		cw := csv.NewWriter(f)
		defer cw.Flush()

		columns := c.columns.Columns().VisibleColumns()

		colNames := make([]string, len(columns))
		for i, c := range columns {
			colNames[i] = c.Name
		}
		if err := cw.Write(colNames); err != nil {
			return 0, errors.Wrapf(err, "cannot export to '%v'", filename)
		}

		totalRows := 0
		row := make([]string, len(columns))
		for {
			for _, item := range resultSet.Items() {
				for i, col := range columns {
					row[i], _ = attrutils.AttributeToString(col.Evaluator.EvaluateForItem(item))
				}
				if err := cw.Write(row); err != nil {
					return 0, errors.Wrapf(err, "cannot export to '%v'", filename)
				}
			}
			totalRows += len(resultSet.Items())

			if !opts.AllResults || !resultSet.HasNextPage() {
				break
			}

			jobs.PostUpdate(ctx, fmt.Sprintf("exported %d items", totalRows))
			resultSet, err = c.tableService.NextPage(ctx, resultSet)
			if err != nil {
				return 0, errors.Wrapf(err, "cannot get next page while exporting to '%v'", filename)
			}
		}

		return totalRows, nil
	}).OnDone(func(rows int) tea.Msg {
		return events.StatusMsg(applyToN("Exported ", rows, "item", "items", " to "+filename))
	}).Submit()
}

type ExportOptions struct {
	// AllResults returns all results from the table
	AllResults bool
}
