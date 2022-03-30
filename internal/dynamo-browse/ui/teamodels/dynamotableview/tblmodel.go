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
)

type itemTableRow struct {
	resultSet *models.ResultSet
	item      models.Item
}

func (mtr itemTableRow) Render(w io.Writer, model table.Model, index int) {
	isMarked := mtr.resultSet.Marked(index)

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
	if index == model.Cursor() {
		style := model.Styles.SelectedRow
		if isMarked {
			style = style.Copy().Inherit(markedRowStyle)
		}
		fmt.Fprintln(w, style.Render(sb.String()))
	} else if isMarked {
		fmt.Fprintln(w, markedRowStyle.Render(sb.String()))
	} else {
		fmt.Fprintln(w, sb.String())
	}
}
