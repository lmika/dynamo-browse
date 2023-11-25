package scriptmanager

import (
	"context"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/queryexpr"
)

type relatedItem struct {
	label string
	query *queryexpr.QueryExpr
}

type relatedItemBuilder struct {
	table          string
	itemProduction func(ctx context.Context, tableInfo *models.TableInfo, item *models.Item) ([]relatedItem, error)
}
