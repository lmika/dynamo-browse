package colselector

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/evaluators"
	table "github.com/lmika/go-bubble-table"
	"io"
)

type colListRowModel struct {
	m *colListModel
}

func (clr colListRowModel) Render(w io.Writer, model table.Model, index int) {
	cols := clr.m.colController.Columns()
	if cols == nil {
		return
	}

	var style lipgloss.Style
	if index == model.Cursor() {
		style = model.Styles.SelectedRow
	}

	col := clr.m.colController.Columns().Columns[index]
	ff := clr.m.sortCriteria.FirstField()
	switch {
	case col.Hidden:
		fmt.Fprintln(w, style.Render(fmt.Sprintf("✕\t%v", col.Name)))
	case evaluators.Equals(ff.Field, col.Evaluator):
		if ff.Asc {
			fmt.Fprintln(w, style.Render(fmt.Sprintf("v\t%v", col.Name)))
		} else {
			fmt.Fprintln(w, style.Render(fmt.Sprintf("^\t%v", col.Name)))
		}
	default:
		fmt.Fprintln(w, style.Render(fmt.Sprintf("⋅\t%v", col.Name)))
	}
}
