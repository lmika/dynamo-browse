package jobs

import (
	"context"
	bus "github.com/lmika/events"
	"sync"
)

type Job func(ctx context.Context)

type jobInfo struct {
	ctx      context.Context
	cancelFn func()
}

type Services struct {
	bus *bus.Bus

	mutex         *sync.Mutex
	foregroundJob *jobInfo
}

func NewService(bus *bus.Bus) *Services {
	return &Services{
		bus:   bus,
		mutex: new(sync.Mutex),
	}
}

// SubmitForegroundJob starts a foreground job.
func (jc *Services) SubmitForegroundJob(job Job, onJobUpdate func(msg string)) {
	// TODO: if there's already a foreground job, then return error

	ctx, cancelFn := context.WithCancel(context.Background())

	jobUpdateChan := make(chan string)
	jobUpdater := &jobUpdaterValue{msgUpdate: jobUpdateChan}
	ctx = context.WithValue(ctx, jobUpdaterKey, jobUpdater)

	newJobInfo := &jobInfo{
		ctx:      ctx,
		cancelFn: cancelFn,
	}
	// TODO: needs to be protected by the mutex
	jc.foregroundJob = newJobInfo

	go func() {
		defer cancelFn()
		defer close(jobUpdateChan)

		job(newJobInfo.ctx)

		// TODO: needs to be protected by the mutex
		jc.foregroundJob = nil
	}()

	go func() {
		for update := range jobUpdateChan {
			onJobUpdate(update)
		}
	}()
}

func (jc *Services) CancelForegroundJob() bool {
	// TODO: needs to be protected by the mutex
	if jc.foregroundJob != nil {
		// A nil cancel for a non-nil foreground job indicates that the cancellation function
		// has been called and the job is in the process of stopping
		if jc.foregroundJob.cancelFn == nil {
			return false
		}

		jc.foregroundJob.cancelFn()
		jc.foregroundJob.cancelFn = nil
		return true
	}

	return false
}
