package controllers

import (
	"bytes"
	"encoding/csv"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/attrutils"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/columns"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services"
	"github.com/pkg/errors"
	"io"
	"os"
)

type ExportController struct {
	state              *State
	columns            *ColumnsController
	pasteboardProvider services.PasteboardProvider
}

func NewExportController(
	state *State,
	columns *ColumnsController,
	pasteboardProvider services.PasteboardProvider,
) *ExportController {
	return &ExportController{state, columns, pasteboardProvider}
}

func (c *ExportController) ExportCSV(filename string) tea.Msg {
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

	columns := c.columns.Columns().VisibleColumns()

	colNames := make([]string, len(columns))
	for i, c := range columns {
		colNames[i] = c.Name
	}
	if err := cw.Write(colNames); err != nil {
		return events.Error(errors.Wrapf(err, "cannot export to '%v'", filename))
	}

	row := make([]string, len(columns))
	for _, item := range resultSet.Items() {
		for i, col := range columns {
			row[i], _ = attrutils.AttributeToString(col.Evaluator.EvaluateForItem(item))
		}
		if err := cw.Write(row); err != nil {
			return events.Error(errors.Wrapf(err, "cannot export to '%v'", filename))
		}
	}

	return nil
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
