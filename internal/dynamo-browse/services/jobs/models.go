package jobs

import "context"

const (
	JobStartEvent  = "jobs.start"
	JobIdleEvent   = "jobs.idle"
	JobUpdateEvent = "jobs.update"
)

type EventData struct {
	Job Job
}

type Job interface {
	Execute(ctx context.Context)
}

type JobFunc func(ctx context.Context)

func (jf JobFunc) Execute(ctx context.Context) {
	jf(ctx)
}

func WithDescription(description string, job Job) Job {
	return DescribableJob{job, description}
}

type DescribableJob struct {
	Job
	Description string
}
