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

func (s *Service) Clone(ctx context.Context, param models.SSMParameter, newName string) error {
	newParam := models.SSMParameter{
		Name:  newName,
		Type:  param.Type,
		Value: param.Value,
	}
	return s.provider.Put(ctx, newParam, false)
}

func (s *Service) Delete(ctx context.Context, param models.SSMParameter) error {
	return s.provider.Delete(ctx, param)
}
