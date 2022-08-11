package serialisable

import (
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"time"
)

type ResultSetSnapshot struct {
	ID        int64 `storm:"id,increment"`
	BackLink  int64 `storm:"index"`
	Time      time.Time
	TableInfo *models.TableInfo
	Query     Query
	Filter    string
}

type Query struct {
	Expression string
}
