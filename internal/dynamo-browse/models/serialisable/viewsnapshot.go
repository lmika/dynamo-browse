package serialisable

import (
	"time"
)

type ViewSnapshot struct {
	ID       int64 `storm:"id,increment"`
	BackLink int64 `storm:"index"`
	ForeLink int64 `storm:"index"`
	Time     time.Time
	Details  ViewSnapshotDetails
}

type ViewSnapshotDetails struct {
	TableName string
	Query     string // TODO: this needs to be the serialised query, not the query expression
	Filter    string
}
