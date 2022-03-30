package ssmparameters

import (
	"context"
	"github.com/lmika/awstools/internal/ssm-browse/models"
)

type Service struct {
	provider SSMProvider
}

func NewService(provider SSMProvider) *Service {
	return &Service{
		provider: provider,
	}
}

func (s *Service) List(ctx context.Context, prefix string) (*models.SSMParameters, error) {
	return s.provider.List(ctx, prefix, 100)
}