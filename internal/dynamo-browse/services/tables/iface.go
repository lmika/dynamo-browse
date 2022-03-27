package tables

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
)

type TableProvider interface {
	ListTables(ctx context.Context) ([]string, error)
	DescribeTable(ctx context.Context, tableName string) (*models.TableInfo, error)
	ScanItems(ctx context.Context, tableName string) ([]models.Item, error)
	DeleteItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) error
	PutItem(ctx context.Context, name string, item models.Item) error
}
