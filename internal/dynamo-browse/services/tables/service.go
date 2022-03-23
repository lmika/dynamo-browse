package tables

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

	// TODO: need to get PKs and SKs from table
	pk, sk := "pk", "sk"

	// Get the columns
	seenColumns := make(map[string]int)
	seenColumns[pk] = 0
	seenColumns[sk] = 1

	for _, result := range results {
		for k := range result {
			if _, isSeen := seenColumns[k]; !isSeen {
				seenColumns[k] = 2
			}
		}
	}

	columns := make([]string, 0, len(seenColumns))
	for k := range seenColumns {
		columns = append(columns, k)
	}
	sort.Slice(columns, func(i, j int) bool {
		if seenColumns[columns[i]] == seenColumns[columns[j]] {
			return columns[i] < columns[j]
		}
		return seenColumns[columns[i]] < seenColumns[columns[j]]
	})

	models.Sort(results, pk, sk)

	return &models.ResultSet{
		Table:   table,
		Columns: columns,
		Items:   results,
	}, nil
}

func (s *Service) Delete(ctx context.Context, name string, item models.Item) error {
	// TODO: do not hardcode keys
	return s.provider.DeleteItem(ctx, name, map[string]types.AttributeValue{
		"pk": item["pk"],
		"sk": item["sk"],
	})
}
