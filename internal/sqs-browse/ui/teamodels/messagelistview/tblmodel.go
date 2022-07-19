package messagelistview

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/sqs-browse/models"
	table "github.com/lmika/go-bubble-table"
	"io"
	"strings"
)

var (
	markedRowStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#e1e1e1", Dark: "#414141"})
	dirtyRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e13131"))
	newRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#31e131"))

	metaInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
)

type itemTableRow struct {
	model       *Model
	messageList *models.MessageList
	itemIndex   int
	item        models.Message
}

func (mtr itemTableRow) Render(w io.Writer, model table.Model, index int) {
	/*
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
		for i, colName := range mtr.resultSet.Columns()[mtr.model.colOffset:] {
			if i > 0 {
				sb.WriteString(style.Render("\t"))
			}

			if r := mtr.item.Renderer(colName); r != nil {
				sb.WriteString(style.Render(r.StringValue()))
				if mi := r.MetaInfo(); mi != "" {
					sb.WriteString(metaInfoStyle.Render(mi))
				}
			} else {
				sb.WriteString(metaInfoStyle.Render("~"))
			}
		}

		fmt.Fprintln(w, sb.String())
	*/
	sb := strings.Builder{}
	sb.WriteString("\t")
	sb.WriteString(mtr.item.Received.Format("2006-01-02 15:04:05.999"))
	sb.WriteString("\t")
	sb.WriteString(string(mtr.item.Body))
}
