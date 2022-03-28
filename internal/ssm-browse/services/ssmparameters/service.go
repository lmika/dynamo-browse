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
	var items []models.SSMParameter
	var nextToken string

	for {
		page, err := s.provider.List(ctx, prefix, nextToken)
		if err != nil {
			return nil, err
		}

		items = append(items, page.Items...)
		nextToken = page.NextToken
		if len(items) >= 50 || nextToken == "" {
			break
		}
	}

	return &models.SSMParameters{Items: items, NextToken: nextToken}, nil
}