package ssmparameters

import (
	"context"
	"github.com/lmika/awstools/internal/ssm-browse/models"
)

type SSMProvider interface {
	List(ctx context.Context, prefix string, nextToken string) (*models.SSMParameters, error)
}
