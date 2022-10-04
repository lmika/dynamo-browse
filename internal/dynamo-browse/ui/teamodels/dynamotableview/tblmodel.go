package dynamotableview

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/audax/internal/dynamo-browse/models/itemrender"
	"io"
	"strings"

	"github.com/lmika/audax/internal/dynamo-browse/models"
	table "github.com/lmika/go-bubble-table"
)

var (
	markedRowStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#e1e1e1", Dark: "#414141"})
	dirtyRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e13131"))
	newRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#2B800C", Dark: "#73C653"})

	metaInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
)

type itemTableRow struct {
	model     *Model
	resultSet *models.ResultSet
	itemIndex int
	item      models.Item
}

func (mtr itemTableRow) Render(w io.Writer, model table.Model, index int) {
	isMarked := mtr.resultSet.Marked(mtr.itemIndex)
	isDirty := mtr.resultSet.IsDirty(mtr.itemIndex)
	isNew := mtr.resultSet.IsNew(mtr.itemIndex)

	var style lipgloss.Style

	if index == model.Cursor() {
		style = model.Styles.SelectedRow
	}
	if isMarked {
		style = style.Copy().Inherit(markedRowStyle)
	}
	if isNew {
		style = style.Copy().Inherit(newRowStyle)
	} else if isDirty {
		style = style.Copy().Inherit(dirtyRowStyle)
	}
	metaInfoStyle := style.Copy().Inherit(metaInfoStyle)

	sb := strings.Builder{}

	// The status column
	switch {
	case isNew:
		sb.WriteString(style.Render("*\t"))
	case isDirty:
		sb.WriteString(style.Render("M\t"))
	case isMarked:
		sb.WriteString(style.Render("•\t"))
	default:
		sb.WriteString(metaInfoStyle.Render("⋅\t"))
	}

	for i, col := range mtr.model.columns[mtr.model.colOffset:] {
		if i > 0 {
			sb.WriteString(style.Render("\t"))
		}

		if r := itemrender.ToRenderer(col.Evaluator.EvaluateForItem(mtr.item)); r != nil {
			sb.WriteString(style.Render(r.StringValue()))
			if mi := r.MetaInfo(); mi != "" {
				sb.WriteString(metaInfoStyle.Render(mi))
			}
		} else {
			sb.WriteString(metaInfoStyle.Render("~"))
		}
	}

	fmt.Fprintln(w, sb.String())
}
