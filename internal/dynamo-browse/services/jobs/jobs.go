package jobs

import (
	"context"
	bus "github.com/lmika/events"
	"github.com/pkg/errors"
	"sync"
)

type jobInfo struct {
	ctx      context.Context
	job      Job
	cancelFn func()
}

type Services struct {
	bus      *bus.Bus
	jobQueue chan Job

	mutex         *sync.Mutex
	foregroundJob *jobInfo
}

func NewService(bus *bus.Bus) *Services {
	jc := &Services{
		bus:      bus,
		jobQueue: make(chan Job, 10),
		mutex:    new(sync.Mutex),
	}
	go jc.waitForJobs()
	return jc
}

// SubmitForegroundJob starts a foreground job.
func (jc *Services) SubmitForegroundJob(job Job) error {
	select {
	case jc.jobQueue <- job:
		return nil
	default:
		return errors.New("too many jobs queued")
	}
}

func (jc *Services) setForegroundJob(newJobInfo *jobInfo) {
	jc.mutex.Lock()
	jc.foregroundJob = newJobInfo
	jc.mutex.Unlock()

	if newJobInfo != nil {
		jc.bus.Fire(JobStartEvent, EventData{Job: newJobInfo.job})
	} else {
		jc.bus.Fire(JobIdleEvent)
	}
}

func (jc *Services) CancelForegroundJob() bool {
	jc.mutex.Lock()
	defer jc.mutex.Unlock()

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

func (jc *Services) waitForJobs() {
	ctx := context.Background()

	for job := range jc.jobQueue {
		jc.runJob(ctx, job)

		if len(jc.jobQueue) == 0 {
			jc.setForegroundJob(nil)
		}
	}
}

func (jc *Services) runJob(ctx context.Context, job Job) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	updateCloseChan := make(chan struct{})
	jobUpdateChan := make(chan string)

	jobUpdater := &jobUpdaterValue{msgUpdate: jobUpdateChan}
	ctx = context.WithValue(ctx, jobUpdaterKey, jobUpdater)

	newJobInfo := &jobInfo{
		job:      job,
		ctx:      ctx,
		cancelFn: cancelFn,
	}
	jc.setForegroundJob(newJobInfo)

	go func() {
		defer close(updateCloseChan)

		for update := range jobUpdateChan {
			jc.bus.Fire(JobUpdateEvent, EventData{Job: job}, update)
		}
	}()

	job.Execute(newJobInfo.ctx)

	close(jobUpdateChan)
	<-updateCloseChan
}
