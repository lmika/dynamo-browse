package dynamoitemedit

import (
	"fmt"
	table "github.com/calyptia/go-bubble-table"
	"io"
)

type itemModel struct {
	model     *Model
	name      string
	attrType  string
	attrValue string
}

func (i itemModel) Render(w io.Writer, model table.Model, index int) {
	var line string
	if i.model.editMode != nil {
		fmt.Fprint(w, i.model.editMode.textInput.View())
		return
		//line = fmt.Sprintf("%s\t%s\t%s", i.name, i.attrType, i.model.editMode.textInput.View())
	} else {
		line = fmt.Sprintf("%s\t%s\t%s", i.name, i.attrType, i.attrValue)
	}

	if index == model.Cursor() {
		fmt.Fprintln(w, model.Styles.SelectedRow.Render(line))
	} else {
		fmt.Fprintln(w, line)
	}
}
