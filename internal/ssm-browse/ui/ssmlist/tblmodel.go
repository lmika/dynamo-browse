package ssmlist

import (
	"fmt"
	table "github.com/calyptia/go-bubble-table"
	"github.com/lmika/awstools/internal/ssm-browse/models"
	"io"
)

type itemTableRow struct {
	item      models.SSMParameter
}

func (mtr itemTableRow) Render(w io.Writer, model table.Model, index int) {
	line := fmt.Sprintf("%s\t%s\t%s", mtr.item.Name, "String", mtr.item.Value)

	if index == model.Cursor() {
		fmt.Fprintln(w, model.Styles.SelectedRow.Render(line))
	} else {
		fmt.Fprintln(w, line)
	}
}
