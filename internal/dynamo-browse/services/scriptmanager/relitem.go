package scriptmanager

import (
	"context"
	"log"
	"path"

	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/queryexpr"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/relitems"
)

type relatedItem struct {
	label string
	table string
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
			log.Printf("RelatedItemOfItem: table = '%v', pattern = '%v', match = '%v'", rb.table, rs.TableInfo.Name, match)
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
						Table: ri.table,
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
