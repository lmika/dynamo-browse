package controllers

import (
	"fmt"
	"github.com/lmika/audax/internal/ssm-browse/models"
)

type NewParameterListMsg struct {
	Prefix string
	Parameters *models.SSMParameters
}

func (rs NewParameterListMsg) StatusMessage() string {
	return fmt.Sprintf("%d items returned", len(rs.Parameters.Items))
}