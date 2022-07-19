package models

import "time"

type Group struct {
	ID         uint64 `storm:"id,increment"`
	Name       string
	Received   time.Time
	CreatedBy  string
	SourceType string
	Source     string
}

type Message struct {
	ID       uint64 `storm:"id,increment"`
	ExtID    string `storm:"unique"`
	GroupID  uint64 `storm:"index"`
	Received time.Time
	Body     string
}

type Event struct {
	ID          uint64    `storm:"id,increment"`
	Date        time.Time `storm:"index"`
	MessageID   uint64    `storm:"index"`
	Operation   string
	QueueName   string
	SuccessCode int
}
