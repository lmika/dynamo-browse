package inputhistorystore

import "time"

type inputHistoryItem struct {
	ID       int    `storm:"id,increment"`
	Category string `storm:"index"`
	Time     time.Time
	Item     string
}
