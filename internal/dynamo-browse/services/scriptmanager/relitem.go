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
	itemProduction func(ctx context.Context, rs *models.ResultSet, index int) ([]relatedItem, error)
}

func (s *Service) RelatedItemOfItem(ctx context.Context, rs *models.ResultSet, index int) ([]models.RelatedItem, error) {
	riModels := []models.RelatedItem{}

	for _, plugin := range s.plugins {
		for _, rb := range plugin.relatedItems {
			// TODO: should support matching
			if rb.table == rs.TableInfo.Name {
				relatedItems, err := rb.itemProduction(ctx, rs, index)
				if err != nil {
					// TODO: should probably return error if no rel items were found and an error was raised
					return nil, err
				}

				// TODO: make this nicer
				for _, ri := range relatedItems {
					riModels = append(riModels, models.RelatedItem{
						Name: ri.label,
					})
				}
			}
		}
	}
	return riModels, nil
}
