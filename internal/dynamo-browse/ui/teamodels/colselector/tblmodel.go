package colselector

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
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
	if !col.Hidden {
		fmt.Fprintln(w, style.Render(fmt.Sprintf(".\t%v", col.Name)))
	} else {
		fmt.Fprintln(w, style.Render(fmt.Sprintf("✕\t%v", col.Name)))
	}
}
