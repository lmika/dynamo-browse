package models

import "time"

type Message struct {
	ID       uint64
	ExtID string
	Queue    string
	Received time.Time
	Data     string
}
