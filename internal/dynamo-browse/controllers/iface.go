package controllers

import (
	"context"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
)

type TableReadService interface {
	ListTables(background context.Context) ([]string, error)
	Describe(ctx context.Context, table string) (*models.TableInfo, error)
	Scan(ctx context.Context, tableInfo *models.TableInfo) (*models.ResultSet, error)
	Filter(resultSet *models.ResultSet, filter string) *models.ResultSet
}
