package jobs

const (
	JobEventForegroundDone = "job_foreground_done"
)

type JobDoneEvent struct {
	Err error
}
