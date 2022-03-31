package controllers

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/ssm-browse/services/ssmparameters"
	"sync"
)

type SSMController struct {
	service *ssmparameters.Service

	// state
	mutex *sync.Mutex
	prefix string
}

func New(service *ssmparameters.Service) *SSMController {
	return &SSMController{
		service: service,
		prefix: "/",
		mutex: new(sync.Mutex),
	}
}

func (c *SSMController) Fetch() tea.Cmd {
	return func() tea.Msg {
		res, err := c.service.List(context.Background(), c.prefix)
		if err != nil {
			return events.Error(err)
		}

		return NewParameterListMsg{
			Prefix: c.prefix,
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
			Prefix: c.prefix,
			Parameters: res,
		}
	}
}