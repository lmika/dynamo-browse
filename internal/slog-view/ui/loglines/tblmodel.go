package loglines

import (
	"fmt"
	"github.com/lmika/audax/internal/slog-view/models"
	table "github.com/lmika/go-bubble-table"
	"io"
	"strings"
)

type column struct {
	field  string
	maxLen int
}

var columns = []column{
	{field: "level", maxLen: 0},
	{field: "subscription_schedule_id", maxLen: 0},
	{field: "error", maxLen: 60},
	{field: "message", maxLen: 0},
}

type itemTableRow struct {
	item models.LogLine
}

func (mtr itemTableRow) Render(w io.Writer, model table.Model, index int) {
	// TODO: these cols are fixed, they should be dynamic
	line := new(strings.Builder)

	for i, col := range columns {
		if i > 0 {
			line.WriteRune('\t')
		}
		line.WriteString(mtr.renderFirstLineOfField(mtr.item.JSON, col))
	}

	if index == model.Cursor() {
		fmt.Fprintln(w, model.Styles.SelectedRow.Render(line.String()))
	} else {
		fmt.Fprintln(w, line.String())
	}
}

// TODO: this needs to be some form of path expression
func (mtr itemTableRow) renderFirstLineOfField(d interface{}, col column) string {
	var singleLine string
	switch k := d.(type) {
	case map[string]interface{}:
		singleLine = mtr.renderFirstLineOfValue(k[col.field])
	default:
		singleLine = mtr.renderFirstLineOfValue(k)
	}

	if col.maxLen > 0 && len(singleLine) > col.maxLen {
		singleLine = singleLine[:col.maxLen-1] + "…"
	}
	return singleLine
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
