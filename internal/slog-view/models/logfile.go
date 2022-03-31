package models

type LogFile struct {
	Filename string
	Lines []LogLine
}

type LogLine struct {
	JSON interface{}
}
