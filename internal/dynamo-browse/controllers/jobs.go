package controllers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/services/jobs"
	bus "github.com/lmika/events"
)

type JobsController struct {
	service   *jobs.Services
	msgSender func(msg tea.Msg)
}

func NewJobsController(service *jobs.Services, bus *bus.Bus) *JobsController {
	jc := &JobsController{
		service: service,
	}

	bus.On(jobs.JobEventForegroundDone, jc.onJobDone)
	return jc
}

func (js *JobsController) SetMessageSender(msgSender func(msg tea.Msg)) {
	js.msgSender = msgSender
}

func (js *JobsController) SubmitJob(job jobs.Job) tea.Msg {
	js.service.SubmitForegroundJob(job)
	return events.ForegroundJobUpdate{
		JobRunning: true,
		JobStatus:  "Job started… press Ctrl+C to stop",
	}
}

func (jc *JobsController) onJobDone(eventData jobs.JobDoneEvent) {
	if err := eventData.Err; err != nil {
		jc.msgSender(events.ForegroundJobUpdate{
			JobRunning: false,
			JobStatus:  "Err: " + err.Error(),
		})
		return
	}

	jc.msgSender(events.ForegroundJobUpdate{
		JobRunning: false,
		JobStatus:  "",
	})
}

func (js *JobsController) CancelRunningJob() tea.Msg {
	hasCancelled := js.service.CancelForegroundJob()
	if !hasCancelled {
		return events.ForegroundJobUpdate{
			JobRunning: true,
			JobStatus:  "Cancelling running job…",
		}
	}
	return events.StatusMsg("No running job")
}
