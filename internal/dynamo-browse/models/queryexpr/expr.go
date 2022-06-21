package queryexpr

import "github.com/lmika/awstools/internal/dynamo-browse/models"

type QueryExpr struct {
	ast *astExpr
}

func (md *QueryExpr) BuildQuery(tableInfo *models.TableInfo) (*models.QueryExecutionPlan, error) {
	return md.ast.calcQuery(tableInfo)
}
