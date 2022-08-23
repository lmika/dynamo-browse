package serialisable

import (
	"time"
)

type ViewSnapshot struct {
	ID        int64 `storm:"id,increment"`
	BackLink  int64 `storm:"index"`
	ForeLink  int64 `storm:"index"`
	Time      time.Time
	TableName string
	Query     string
	Filter    string
}
