package ssmparameters

import (
	"context"
	"github.com/lmika/audax/internal/ssm-browse/models"
)

type SSMProvider interface {
	List(ctx context.Context, prefix string, maxCount int) (*models.SSMParameters, error)
	Put(ctx context.Context, param models.SSMParameter, override bool) error
	Delete(ctx context.Context, param models.SSMParameter) error
}
