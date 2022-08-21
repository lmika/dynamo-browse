package serialisable

import (
	"time"
)

type ViewSnapshot struct {
	ID        int64 `storm:"id,increment"`
	BackLink  int64 `storm:"index"`
	Time      time.Time
	TableName string
	Query     string
	Filter    string
}

func (vs *ViewSnapshot) IsSameView(other *ViewSnapshot) bool {
	return vs.TableName == other.TableName && vs.Query == other.Query && vs.Filter == other.Filter
}
