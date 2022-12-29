package scriptmanager

import (
	"context"
	"github.com/pkg/errors"
)

type scriptScheduler struct {
	jobChan chan scriptJob
}

func newScriptScheduler() *scriptScheduler {
	ss := &scriptScheduler{}
	ss.start()
	return ss
}

func (ss *scriptScheduler) start() {
	ss.jobChan = make(chan scriptJob)
	go func() {
		for job := range ss.jobChan {
			job.job(job.ctx)
		}
	}()
}

// startJobOnceFree will submit a script execution job.  The function will wait until the scheduler is free.
// The job will then run on the script goroutine and the function will return.
func (ss *scriptScheduler) startJobOnceFree(ctx context.Context, job func(ctx context.Context)) error {
	select {
	case ss.jobChan <- scriptJob{ctx: ctx, job: job}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// runNow will submit a job for immediate execution.  The job will run as long as the scheduler is free.
// If the scheduler is not free, an error will be returned and the job will not run.
func (ss *scriptScheduler) runNow(ctx context.Context, job func(ctx context.Context)) error {
	select {
	case ss.jobChan <- scriptJob{ctx: ctx, job: job}:
		return nil
	default:
		return errors.New("a script is already running")
	}
}

type scriptJob struct {
	ctx context.Context
	job func(ctx context.Context)
}
