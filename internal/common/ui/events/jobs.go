package events

type ForegroundJobUpdate struct {
	JobRunning bool
	JobStatus  string
}
