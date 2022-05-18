package ssmlist

import (
	"fmt"
	table "github.com/calyptia/go-bubble-table"
	"github.com/lmika/awstools/internal/ssm-browse/models"
	"io"
	"strings"
)

type itemTableRow struct {
	item      models.SSMParameter
}

func (mtr itemTableRow) Render(w io.Writer, model table.Model, index int) {
	firstLine := strings.SplitN(mtr.item.Value, "\n", 2)[0]
	line := fmt.Sprintf("%s\t%s\t%s", mtr.item.Name, mtr.item.Type, firstLine)

	if index == model.Cursor() {
		fmt.Fprintln(w, model.Styles.SelectedRow.Render(line))
	} else {
		fmt.Fprintln(w, line)
	}
}
