package tables

import (
	"context"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"sort"
)

type Service struct {
	provider TableProvider
}

func NewService(provider TableProvider) *Service {
	return &Service{
		provider: provider,
	}
}

func (s *Service) Scan(ctx context.Context, table string) (*models.ResultSet, error) {
	results, err := s.provider.ScanItems(ctx, table)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to scan table %v", table)
	}

	// Get the columns
	// TODO: need to get PKs and SKs from table
	seenColumns := make(map[string]int)
	seenColumns["pk"] = 0
	seenColumns["sk"] = 1

	for _, result := range results {
		for k := range result {
			if _, isSeen := seenColumns[k]; !isSeen {
				seenColumns[k] = len(seenColumns)
			}
		}
	}

	columns := make([]string, 0, len(seenColumns))
	for k := range seenColumns {
		columns = append(columns, k)
	}
	sort.Slice(columns, func(i, j int) bool {
		return seenColumns[columns[i]] < seenColumns[columns[j]]
	})

	return &models.ResultSet{
		Columns: columns,
		Items: results,
	}, nil
}
