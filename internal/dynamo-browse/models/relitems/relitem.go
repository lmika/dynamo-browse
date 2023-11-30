package relitems

import "github.com/lmika/dynamo-browse/internal/dynamo-browse/models/queryexpr"

type RelatedItem struct {
	Name  string
	Query *queryexpr.QueryExpr
}
