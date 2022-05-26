package dynamotableview

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	table "github.com/calyptia/go-bubble-table"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
)

var (
	markedRowStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#e1e1e1"))
	dirtyRowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e13131"))
	newRowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#31e131"))
)

type itemTableRow struct {
	resultSet *models.ResultSet
	itemIndex int
	item      models.Item
}

func (mtr itemTableRow) Render(w io.Writer, model table.Model, index int) {
	isMarked := mtr.resultSet.Marked(mtr.itemIndex)
	isDirty := mtr.resultSet.IsDirty(mtr.itemIndex)
	isNew := mtr.resultSet.IsNew(mtr.itemIndex)

	sb := strings.Builder{}
	for i, colName := range mtr.resultSet.Columns {
		if i > 0 {
			sb.WriteString("\t")
		}

		switch colVal := mtr.item[colName].(type) {
		case nil:
			sb.WriteString("(nil)")
		case *types.AttributeValueMemberS:
			sb.WriteString(colVal.Value)
		case *types.AttributeValueMemberN:
			sb.WriteString(colVal.Value)
		default:
			sb.WriteString("(other)")
		}
	}

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

	fmt.Fprintln(w, style.Render(sb.String()))
}
