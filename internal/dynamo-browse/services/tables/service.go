package tables

import (
	"context"
	"sort"

	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

type Service struct {
	provider TableProvider
}

func NewService(provider TableProvider) *Service {
	return &Service{
		provider: provider,
	}
}

func (s *Service) ListTables(ctx context.Context) ([]string, error) {
	return s.provider.ListTables(ctx)
}

func (s *Service) Describe(ctx context.Context, table string) (*models.TableInfo, error) {
	return s.provider.DescribeTable(ctx, table)
}

func (s *Service) Scan(ctx context.Context, tableInfo *models.TableInfo) (*models.ResultSet, error) {
	results, err := s.provider.ScanItems(ctx, tableInfo.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to scan table %v", tableInfo.Name)
	}

	// Get the columns
	seenColumns := make(map[string]int)
	seenColumns[tableInfo.Keys.PartitionKey] = 0
	if tableInfo.Keys.SortKey != "" {
		seenColumns[tableInfo.Keys.SortKey] = 1
	}

	for _, definedAttribute := range tableInfo.DefinedAttributes {
		if _, seen := seenColumns[definedAttribute]; !seen {
			seenColumns[definedAttribute] = len(seenColumns)
		}
	}

	otherColsRank := len(seenColumns)
	for _, result := range results {
		for k := range result {
			if _, isSeen := seenColumns[k]; !isSeen {
				seenColumns[k] = otherColsRank
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

	models.Sort(results, tableInfo)

	return &models.ResultSet{
		TableInfo: tableInfo,
		Columns:   columns,
		Items:     results,
	}, nil
}

func (s *Service) Put(ctx context.Context, tableInfo *models.TableInfo, item models.Item) error {
	return s.provider.PutItem(ctx, tableInfo.Name, item)
}

func (s *Service) Delete(ctx context.Context, tableInfo *models.TableInfo, items []models.Item) error {
	for _, item := range items {
		if err := s.provider.DeleteItem(ctx, tableInfo.Name, item.KeyValue(tableInfo)); err != nil {
			return errors.Wrapf(err, "cannot delete item")
		}
	}
	return nil
}
