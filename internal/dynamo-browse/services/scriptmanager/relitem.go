package scriptmanager

import (
	"context"
	"path"

	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/queryexpr"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/relitems"
)

type relatedItem struct {
	label string
	query *queryexpr.QueryExpr
}

type relatedItemBuilder struct {
	table          string
	itemProduction func(ctx context.Context, rs *models.ResultSet, index int) ([]relatedItem, error)
}

func (s *Service) RelatedItemOfItem(ctx context.Context, rs *models.ResultSet, index int) ([]relitems.RelatedItem, error) {
	riModels := []relitems.RelatedItem{}

	for _, plugin := range s.plugins {
		for _, rb := range plugin.relatedItems {
			// TODO: should support matching
			match, _ := tableMatchesGlob(rb.table, rs.TableInfo.Name)
			if match {
				relatedItems, err := rb.itemProduction(ctx, rs, index)
				if err != nil {
					// TODO: should probably return error if no rel items were found and an error was raised
					return nil, err
				}

				// TODO: make this nicer
				for _, ri := range relatedItems {
					riModels = append(riModels, relitems.RelatedItem{
						Name:  ri.label,
						Query: ri.query,
					})
				}
			}
		}
	}
	return riModels, nil
}

func tableMatchesGlob(tableName, pattern string) (bool, error) {
	return path.Match(tableName, pattern)
}
