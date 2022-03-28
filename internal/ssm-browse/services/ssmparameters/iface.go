package ssmparameters

import (
	"context"
	"github.com/lmika/awstools/internal/ssm-browse/models"
)

type SSMProvider interface {
	List(ctx context.Context) (*models.SSMParameters, error)
}
