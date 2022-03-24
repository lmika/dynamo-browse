package models

import "time"

type Message struct {
	ID       uint64		`storm:"id,increment"`
	ExtID    string		`storm:"unique"`
	Queue    string		`storm:"index"`
	Received time.Time
	Data     string
}
