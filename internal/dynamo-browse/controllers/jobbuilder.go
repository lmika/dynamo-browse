package controllers

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/jobs"
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
		if jb.onDone == nil {
			return nil
		}
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
	if err := jb.jc.service.SubmitForegroundJob(jobs.WithDescription(jb.description, jobs.JobFunc(func(ctx context.Context) {
		msg := jb.executeJob(ctx)

		jb.jc.msgSender(msg)

		if _, isForegroundJobUpdate := msg.(events.ForegroundJobUpdate); !isForegroundJobUpdate {
			// Likely another job was scheduled so don't indicate that no jobs are running.
			jb.jc.msgSender(events.ForegroundJobUpdate{
				JobRunning: false,
				JobStatus:  "",
			})
		}
	}))); err != nil {
		return events.Error(err)
	}

	return events.ForegroundJobUpdate{
		JobRunning: true,
		JobStatus:  jb.description,
	}
}
