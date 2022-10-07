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
	onEither    func(res T, err error) tea.Msg
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

func (jb JobBuilder[T]) OnEither(fn func(res T, err error) tea.Msg) JobBuilder[T] {
	newJb := jb
	newJb.onEither = fn
	return newJb
}

func (jb JobBuilder[T]) Submit() tea.Msg {
	if jb.jc.immediate {
		return jb.executeJob(context.Background())
	}
	return jb.doSubmit()
}

func (jb JobBuilder[T]) executeJob(ctx context.Context) tea.Msg {
	res, err := jb.job(ctx)

	if jb.onEither != nil {
		return jb.onEither(res, err)
	} else if err == nil {
		return jb.onDone(res)
	} else {
		if jb.onErr != nil {
			return jb.onErr(err)
		} else {
			return events.Error(err)
		}
	}
}

func (jb JobBuilder[T]) doSubmit() tea.Msg {
	jobFinished := make(chan tea.Msg)

	jb.jc.service.SubmitForegroundJob(func(ctx context.Context) {
		msg := jb.executeJob(ctx)

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
	}, func(msg string) {
		jb.jc.msgSender(events.ForegroundJobUpdate{
			JobRunning: true,
			JobStatus:  jb.description + " " + msg,
		})
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
