package tables

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/common/sliceutils"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/jobs"
	"log"
	"strings"
	"time"

	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

type Service struct {
	provider       TableProvider
	configProvider ConfigProvider
}

func NewService(provider TableProvider, roProvider ConfigProvider) *Service {
	return &Service{
		provider:       provider,
		configProvider: roProvider,
	}
}

func (s *Service) ListTables(ctx context.Context) ([]string, error) {
	return s.provider.ListTables(ctx)
}

func (s *Service) Describe(ctx context.Context, table string) (*models.TableInfo, error) {
	return s.provider.DescribeTable(ctx, table)
}

func (s *Service) Scan(ctx context.Context, tableInfo *models.TableInfo) (*models.ResultSet, error) {
	return s.doScan(ctx, tableInfo, nil, nil, s.configProvider.DefaultLimit())
}

func (s *Service) doScan(
	ctx context.Context,
	tableInfo *models.TableInfo,
	expr models.Queryable,
	exclusiveStartKey map[string]types.AttributeValue,
	limit int,
) (*models.ResultSet, error) {
	var (
		filterExpr *expression.Expression
		runAsQuery bool
		index      string
		err        error
	)
	if expr != nil {
		plan, err := expr.Plan(tableInfo)
		if err != nil {
			return nil, err
		}

		runAsQuery = plan.CanQuery
		index = plan.IndexName
		filterExpr = &plan.Expression

		log.Printf("Running query over '%v'", tableInfo.Name)
		plan.Describe(log.Default())
	} else {
		log.Printf("Performing scan over '%v'", tableInfo.Name)
	}

	var results []models.Item
	var lastEvalKey map[string]types.AttributeValue
	if runAsQuery {
		results, lastEvalKey, err = s.provider.QueryItems(ctx, tableInfo.Name, index, filterExpr, exclusiveStartKey, limit)
	} else {
		results, lastEvalKey, err = s.provider.ScanItems(ctx, tableInfo.Name, filterExpr, exclusiveStartKey, limit)
	}

	if err != nil && len(results) == 0 {
		return &models.ResultSet{
			TableInfo:         tableInfo,
			Query:             expr,
			ExclusiveStartKey: exclusiveStartKey,
			LastEvaluatedKey:  lastEvalKey,
		}, errors.Wrapf(err, "unable to scan table %v", tableInfo.Name)
	}

	models.Sort(results, tableInfo)

	resultSet := &models.ResultSet{
		TableInfo:         tableInfo,
		Query:             expr,
		ExclusiveStartKey: exclusiveStartKey,
		LastEvaluatedKey:  lastEvalKey,
	}
	resultSet.SetItems(results)
	resultSet.RefreshColumns()

	return resultSet, err
}

func (s *Service) Put(ctx context.Context, tableInfo *models.TableInfo, item models.Item) error {
	if err := s.assertReadWrite(); err != nil {
		return err
	}

	return s.provider.PutItem(ctx, tableInfo.Name, item)
}

func (s *Service) PutItemAt(ctx context.Context, resultSet *models.ResultSet, index int) error {
	if err := s.assertReadWrite(); err != nil {
		return err
	}

	item := resultSet.Items()[index]
	if err := s.provider.PutItem(ctx, resultSet.TableInfo.Name, item); err != nil {
		return err
	}

	resultSet.SetDirty(index, false)
	resultSet.SetNew(index, false)
	return nil
}

func (s *Service) PutSelectedItems(ctx context.Context, resultSet *models.ResultSet, markedItems []models.ItemIndex) error {
	if err := s.assertReadWrite(); err != nil {
		return err
	}

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
	if err := s.assertReadWrite(); err != nil {
		return err
	}

	nextUpdate := time.Now().Add(1 * time.Second)

	for i, item := range items {
		if err := s.provider.DeleteItem(ctx, tableInfo.Name, item.KeyValue(tableInfo)); err != nil {
			return errors.Wrapf(err, "cannot delete item")
		}

		if time.Now().After(nextUpdate) {
			jobs.PostUpdate(ctx, fmt.Sprintf("delete %d items", i))
			nextUpdate = time.Now().Add(1 * time.Second)
		}
	}
	return nil
}

func (s *Service) ScanOrQuery(ctx context.Context, tableInfo *models.TableInfo, expr models.Queryable, exclusiveStartKey map[string]types.AttributeValue) (*models.ResultSet, error) {
	return s.doScan(ctx, tableInfo, expr, exclusiveStartKey, s.configProvider.DefaultLimit())
}

func (s *Service) NextPage(ctx context.Context, resultSet *models.ResultSet) (*models.ResultSet, error) {
	return s.doScan(ctx, resultSet.TableInfo, resultSet.Query, resultSet.LastEvaluatedKey, s.configProvider.DefaultLimit())
}

func (s *Service) assertReadWrite() error {
	b, err := s.configProvider.IsReadOnly()
	if err != nil {
		return err
	} else if b {
		return models.ErrReadOnly
	}
	return nil
}

// TODO: move into a new service
func (s *Service) Filter(resultSet *models.ResultSet, filter string) *models.ResultSet {
	if resultSet == nil {
		return nil
	}

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
