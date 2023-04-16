package controllers

import (
	"encoding/csv"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/attrutils"
	"github.com/pkg/errors"
	"os"
)

type ExportController struct {
	state   *State
	columns *ColumnsController
}

func NewExportController(state *State, columns *ColumnsController) *ExportController {
	return &ExportController{state, columns}
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
