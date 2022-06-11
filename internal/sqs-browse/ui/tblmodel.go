package ui

import (
	"fmt"
	"io"
	"strings"

	table "github.com/lmika/go-bubble-table"
	"github.com/lmika/awstools/internal/sqs-browse/models"
)

type messageTableRow models.Message

func (mtr messageTableRow) Render(w io.Writer, model table.Model, index int) {
	firstLine := strings.SplitN(string(mtr.Data), "\n", 2)[0]

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%d", mtr.ID))
	sb.WriteString("\t")
	sb.WriteString(firstLine)

	if index == model.Cursor() {
		fmt.Fprintln(w, model.Styles.SelectedRow.Render(sb.String()))
	} else {
		fmt.Fprintln(w, sb.String())
	}
}
