package controllers

import "github.com/lmika/awstools/internal/ssm-browse/models"

type NewParameterListMsg struct {
	Parameters *models.SSMParameters
}
