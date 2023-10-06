package controllers

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/attrutils"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/columns"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/jobs"
	"github.com/pkg/errors"
)

type ExportController struct {
	state              *State
	tableService       TableReadService
	jobController      *JobsController
	columns            *ColumnsController
	pasteboardProvider services.PasteboardProvider
}

func NewExportController(
	state *State,
	tableService TableReadService,
	jobsController *JobsController,
	columns *ColumnsController,
	pasteboardProvider services.PasteboardProvider,
) *ExportController {
	return &ExportController{state, tableService, jobsController, columns, pasteboardProvider}
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

func (c *ExportController) ExportCSVToClipboard() tea.Msg {
	var bts bytes.Buffer

	resultSet := c.state.ResultSet()
	if resultSet == nil {
		return errors.New("no result set")
	}

	if err := c.exportCSV(&bts, c.columns.Columns().VisibleColumns(), resultSet); err != nil {
		return events.Error(err)
	}

	if err := c.pasteboardProvider.WriteText(bts.Bytes()); err != nil {
		return events.Error(err)
	}
	return nil
}

// TODO: this really needs to be a service!
func (c *ExportController) ExportToWriter(w io.Writer, resultSet *models.ResultSet) error {
	return c.exportCSV(w, columns.NewColumnsFromResultSet(resultSet).Columns, resultSet)
}

func (c *ExportController) exportCSV(w io.Writer, cols []columns.Column, resultSet *models.ResultSet) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	colNames := make([]string, len(cols))
	for i, c := range cols {
		colNames[i] = c.Name
	}
	if err := cw.Write(colNames); err != nil {
		return errors.Wrap(err, "cannot export to clipboard")
	}

	row := make([]string, len(cols))
	for _, item := range resultSet.Items() {
		for i, col := range cols {
			row[i], _ = attrutils.AttributeToString(col.Evaluator.EvaluateForItem(item))
		}
		if err := cw.Write(row); err != nil {
			return errors.Wrap(err, "cannot export to clipboard")
		}
	}

	return nil
}

type ExportOptions struct {
	// AllResults returns all results from the table
	AllResults bool
}
