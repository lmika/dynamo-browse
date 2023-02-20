package controllers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/services/jobs"
	bus "github.com/lmika/events"
	"log"
)

type JobsController struct {
	service   *jobs.Services
	msgSender func(msg tea.Msg)
	immediate bool
}

func NewJobsController(service *jobs.Services, bus *bus.Bus, immediate bool) *JobsController {
	jc := &JobsController{
		service:   service,
		immediate: immediate,
	}
	bus.On(jobs.JobStartEvent, func(job jobs.EventData) { jc.sendForegroundJobState(job.Job, "") })
	bus.On(jobs.JobIdleEvent, func() { jc.sendForegroundJobState(nil, "") })
	bus.On(jobs.JobUpdateEvent, func(job jobs.EventData, update string) { jc.sendForegroundJobState(job.Job, update) })

	return jc
}

func (js *JobsController) SetMessageSender(msgSender func(msg tea.Msg)) {
	js.msgSender = msgSender
}

func (js *JobsController) CancelRunningJob(ifNoJobsRunning func() tea.Msg) tea.Msg {
	hasCancelled := js.service.CancelForegroundJob()
	if hasCancelled {
		return events.ForegroundJobUpdate{
			JobRunning: true,
			JobStatus:  "Cancelling running job…",
		}
	}
	return ifNoJobsRunning()
}

func (jc *JobsController) sendForegroundJobState(job jobs.Job, update string) {
	if job == nil {
		log.Printf("job service idle")
		jc.msgSender(events.ForegroundJobUpdate{
			JobRunning: false,
		})
		return
	}

	var statusMessage string
	if dj, ok := job.(jobs.DescribableJob); ok {
		statusMessage = dj.Description
	} else {
		statusMessage = "Working…"
	}

	if len(update) > 0 {
		statusMessage += " " + update
	}
	log.Printf("job update: %v", statusMessage)

	jc.msgSender(events.ForegroundJobUpdate{
		JobRunning: true,
		JobStatus:  statusMessage,
	})
}
