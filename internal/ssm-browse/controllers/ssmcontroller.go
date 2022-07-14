package controllers

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/ssm-browse/models"
	"github.com/lmika/awstools/internal/ssm-browse/services/ssmparameters"
	"sync"
)

type SSMController struct {
	service *ssmparameters.Service

	// state
	mutex  *sync.Mutex
	prefix string
}

func New(service *ssmparameters.Service) *SSMController {
	return &SSMController{
		service: service,
		prefix:  "/",
		mutex:   new(sync.Mutex),
	}
}

func (c *SSMController) Fetch() tea.Cmd {
	return func() tea.Msg {
		res, err := c.service.List(context.Background(), c.prefix)
		if err != nil {
			return events.Error(err)
		}

		return NewParameterListMsg{
			Prefix:     c.prefix,
			Parameters: res,
		}
	}
}

func (c *SSMController) ChangePrefix(newPrefix string) tea.Cmd {
	return func() tea.Msg {
		res, err := c.service.List(context.Background(), newPrefix)
		if err != nil {
			return events.Error(err)
		}

		c.mutex.Lock()
		defer c.mutex.Unlock()
		c.prefix = newPrefix

		return NewParameterListMsg{
			Prefix:     c.prefix,
			Parameters: res,
		}
	}
}

func (c *SSMController) Clone(param models.SSMParameter) tea.Cmd {
	return events.PromptForInput("New key: ", func(value string) tea.Cmd {
		return func() tea.Msg {
			ctx := context.Background()
			if err := c.service.Clone(ctx, param, value); err != nil {
				return events.Error(err)
			}

			res, err := c.service.List(context.Background(), c.prefix)
			if err != nil {
				return events.Error(err)
			}

			return NewParameterListMsg{
				Prefix:     c.prefix,
				Parameters: res,
			}
		}
	})
}

func (c *SSMController) DeleteParameter(param models.SSMParameter) tea.Cmd {
	return events.Confirm("delete parameter? ", func() tea.Cmd {
		return func() tea.Msg {
			ctx := context.Background()
			if err := c.service.Delete(ctx, param); err != nil {
				return events.Error(err)
			}

			res, err := c.service.List(context.Background(), c.prefix)
			if err != nil {
				return events.Error(err)
			}

			return NewParameterListMsg{
				Prefix:     c.prefix,
				Parameters: res,
			}
		}
	})
}
