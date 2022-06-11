package loglines

import (
	"fmt"
	table "github.com/lmika/go-bubble-table"
	"github.com/lmika/awstools/internal/slog-view/models"
	"io"
	"strings"
)

type itemTableRow struct {
	item      models.LogLine
}

func (mtr itemTableRow) Render(w io.Writer, model table.Model, index int) {
	// TODO: these cols are fixed, they should be dynamic
	level := mtr.renderFirstLineOfField(mtr.item.JSON, "level")
	err := mtr.renderFirstLineOfField(mtr.item.JSON, "error")
	msg := mtr.renderFirstLineOfField(mtr.item.JSON, "message")
	line := fmt.Sprintf("%s\t%s\t%s", level, err, msg)

	if index == model.Cursor() {
		fmt.Fprintln(w, model.Styles.SelectedRow.Render(line))
	} else {
		fmt.Fprintln(w, line)
	}
}

// TODO: this needs to be some form of path expression
func (mtr itemTableRow) renderFirstLineOfField(d interface{}, field string) string {
	switch k := d.(type) {
	case map[string]interface{}:
		return mtr.renderFirstLineOfValue(k[field])
	default:
		return mtr.renderFirstLineOfValue(k)
	}
}

func (mtr itemTableRow) renderFirstLineOfValue(v interface{}) string {
	if v == nil {
		return ""
	}

	switch k := v.(type) {
	case string:
		firstLine := strings.SplitN(k, "\n", 2)[0]
		return firstLine
	case int:
		return fmt.Sprint(k)
	case float64:
		return fmt.Sprint(k)
	case bool:
		return fmt.Sprint(k)
	case map[string]interface{}:
		return "{}"
	case []interface{}:
		return "[]"
	default:
		return "(other)"
	}
}