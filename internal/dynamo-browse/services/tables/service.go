package tables

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/awstools/internal/common/sliceutils"
	"strings"

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
	return s.doScan(ctx, tableInfo, nil)
}

func (s *Service) doScan(ctx context.Context, tableInfo *models.TableInfo, expr models.Queryable) (*models.ResultSet, error) {
	var filterExpr *expression.Expression

	if expr != nil {
		plan, err := expr.Plan(tableInfo)
		if err != nil {
			return nil, err
		}

		// TEMP
		if plan.CanQuery {
			return nil, errors.Errorf("queries not yet supported")
		}

		filterExpr = &plan.Expression
	}

	results, err := s.provider.ScanItems(ctx, tableInfo.Name, filterExpr, 1000)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to scan table %v", tableInfo.Name)
	}

	// Get the columns
	//seenColumns := make(map[string]int)
	//seenColumns[tableInfo.Keys.PartitionKey] = 0
	//if tableInfo.Keys.SortKey != "" {
	//	seenColumns[tableInfo.Keys.SortKey] = 1
	//}
	//
	//for _, definedAttribute := range tableInfo.DefinedAttributes {
	//	if _, seen := seenColumns[definedAttribute]; !seen {
	//		seenColumns[definedAttribute] = len(seenColumns)
	//	}
	//}
	//
	//otherColsRank := len(seenColumns)
	//for _, result := range results {
	//	for k := range result {
	//		if _, isSeen := seenColumns[k]; !isSeen {
	//			seenColumns[k] = otherColsRank
	//		}
	//	}
	//}
	//
	//columns := make([]string, 0, len(seenColumns))
	//for k := range seenColumns {
	//	columns = append(columns, k)
	//}
	//sort.Slice(columns, func(i, j int) bool {
	//	if seenColumns[columns[i]] == seenColumns[columns[j]] {
	//		return columns[i] < columns[j]
	//	}
	//	return seenColumns[columns[i]] < seenColumns[columns[j]]
	//})

	models.Sort(results, tableInfo)

	resultSet := &models.ResultSet{
		TableInfo: tableInfo,
		Query:     expr,
		//Columns:   columns,
	}
	resultSet.SetItems(results)
	resultSet.RefreshColumns()

	return resultSet, nil
}

func (s *Service) Put(ctx context.Context, tableInfo *models.TableInfo, item models.Item) error {
	return s.provider.PutItem(ctx, tableInfo.Name, item)
}

func (s *Service) PutItemAt(ctx context.Context, resultSet *models.ResultSet, index int) error {
	item := resultSet.Items()[index]
	if err := s.provider.PutItem(ctx, resultSet.TableInfo.Name, item); err != nil {
		return err
	}

	resultSet.SetDirty(index, false)
	resultSet.SetNew(index, false)
	return nil
}

func (s *Service) PutSelectedItems(ctx context.Context, resultSet *models.ResultSet, markedItems []models.ItemIndex) error {
	if len(markedItems) == 0 {
		return nil
	}

	if err := s.provider.PutItems(ctx, resultSet.TableInfo.Name, sliceutils.Map(markedItems, func(t models.ItemIndex) models.Item {
		return t.Item
	})); err != nil {
		return err
	}

	for _, di := range markedItems {
		resultSet.SetDirty(di.Index, false)
		resultSet.SetNew(di.Index, false)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, tableInfo *models.TableInfo, items []models.Item) error {
	for _, item := range items {
		if err := s.provider.DeleteItem(ctx, tableInfo.Name, item.KeyValue(tableInfo)); err != nil {
			return errors.Wrapf(err, "cannot delete item")
		}
	}
	return nil
}

func (s *Service) ScanOrQuery(ctx context.Context, tableInfo *models.TableInfo, expr models.Queryable) (*models.ResultSet, error) {
	return s.doScan(ctx, tableInfo, expr)
}

// TODO: move into a new service
func (s *Service) Filter(resultSet *models.ResultSet, filter string) *models.ResultSet {
	for i, item := range resultSet.Items() {
		if filter == "" {
			resultSet.SetHidden(i, false)
			continue
		}

		var shouldHide = true
		for k := range item {
			str, ok := item.AttributeValueAsString(k)
			if !ok {
				continue
			}

			if strings.Contains(str, filter) {
				shouldHide = false
				break
			}
		}

		resultSet.SetHidden(i, shouldHide)
	}

	return resultSet
}
