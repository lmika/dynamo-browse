package controllers

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/events"
	"time"
)

func NewJob[T any](jc *JobsController, description string, job func(ctx context.Context) (T, error)) JobBuilder[T] {
	return JobBuilder[T]{jc: jc, description: description, job: job}
}

type JobBuilder[T any] struct {
	jc          *JobsController
	description string
	job         func(ctx context.Context) (T, error)
	onDone      func(res T) tea.Msg
	onErr       func(err error) tea.Msg
}

func (jb JobBuilder[T]) OnDone(fn func(res T) tea.Msg) JobBuilder[T] {
	newJb := jb
	newJb.onDone = fn
	return newJb
}

func (jb JobBuilder[T]) OnErr(fn func(err error) tea.Msg) JobBuilder[T] {
	newJb := jb
	newJb.onErr = fn
	return newJb
}

func (jb JobBuilder[T]) Submit() tea.Msg {
	jobFinished := make(chan tea.Msg)

	jb.jc.service.SubmitForegroundJob(func(ctx context.Context) {
		res, err := jb.job(ctx)

		var msg tea.Msg
		if err == nil {
			msg = jb.onDone(res)
		} else {
			if jb.onErr != nil {
				msg = jb.onErr(err)
			} else {
				msg = events.Error(err)
			}
		}

		select {
		case jobFinished <- msg:
			// Waited for the job to finish, so get it now
		default:
			// Not waiting for job to finish, so
			jb.jc.msgSender(msg)
			jb.jc.msgSender(events.ForegroundJobUpdate{
				JobRunning: false,
				JobStatus:  "",
			})
		}
	})

	// Wait 700 msecs to see if the job finishes in time, otherwise indicate that a job is running
	select {
	case msg := <-jobFinished:
		return msg
	case <-time.After(70 * time.Millisecond):
	}

	return events.ForegroundJobUpdate{
		JobRunning: true,
		JobStatus:  jb.description,
	}
}
