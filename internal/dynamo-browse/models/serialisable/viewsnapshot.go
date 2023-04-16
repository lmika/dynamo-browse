package serialisable

import (
	"bytes"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/queryexpr"
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
	TableName         string
	Query             []byte
	QueryHash         uint64
	Filter            string
	ExclusiveStartKey []byte
}

func (d ViewSnapshotDetails) Equals(other ViewSnapshotDetails, compareHashesOnly bool) bool {
	return d.TableName == other.TableName &&
		d.Filter == other.Filter &&
		bytes.Equal(d.ExclusiveStartKey, d.ExclusiveStartKey) &&
		d.compareQueries(other, compareHashesOnly)
}

func (d ViewSnapshotDetails) compareQueries(other ViewSnapshotDetails, compareHashesOnly bool) bool {
	if d.QueryHash != other.QueryHash {
		return false
	}
	if compareHashesOnly {
		return true
	}

	expr1, err := queryexpr.DeserializeFrom(bytes.NewReader(d.Query))
	if err != nil {
		return false
	}
	expr2, err := queryexpr.DeserializeFrom(bytes.NewReader(other.Query))
	if err != nil {
		return false
	}
	return expr1.Equal(expr2)
}
