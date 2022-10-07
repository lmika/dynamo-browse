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
	immediate bool
}

func NewJobsController(service *jobs.Services, bus *bus.Bus, immediate bool) *JobsController {
	jc := &JobsController{
		service:   service,
		immediate: immediate,
	}

	return jc
}

func (js *JobsController) SetMessageSender(msgSender func(msg tea.Msg)) {
	js.msgSender = msgSender
}

func (js *JobsController) CancelRunningJob() tea.Msg {
	hasCancelled := js.service.CancelForegroundJob()
	if hasCancelled {
		return events.ForegroundJobUpdate{
			JobRunning: true,
			JobStatus:  "Cancelling running job…",
		}
	}
	return events.StatusMsg("No running job")
}
