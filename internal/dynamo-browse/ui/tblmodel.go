package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	table "github.com/calyptia/go-bubble-table"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
)

type itemTableRow struct {
	resultSet *models.ResultSet
	item      models.Item
}

func (mtr itemTableRow) Render(w io.Writer, model table.Model, index int) {
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
		fmt.Fprintln(w, model.Styles.SelectedRow.Render(sb.String()))
	} else {
		fmt.Fprintln(w, sb.String())
	}
}
