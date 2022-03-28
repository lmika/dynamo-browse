package controllers

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/ssm-browse/services/ssmparameters"
)

type SSMController struct {
	service *ssmparameters.Service
}

func New(service *ssmparameters.Service) *SSMController {
	return &SSMController{
		service: service,
	}
}

func (c *SSMController) Fetch() tea.Cmd {
	return func() tea.Msg {
		res, err := c.service.List(context.Background())
		if err != nil {
			return events.Error(err)
		}

		return NewParameterListMsg{
			Parameters: res,
		}
	}
}
