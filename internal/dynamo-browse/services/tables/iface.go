package tables

import (
	"context"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
)

type TableProvider interface {
	ScanItems(ctx context.Context, tableName string) ([]models.Item, error)
}
