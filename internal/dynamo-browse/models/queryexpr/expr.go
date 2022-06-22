package queryexpr

import "github.com/lmika/awstools/internal/dynamo-browse/models"

type QueryExpr struct {
	ast *astExpr
}

func (md *QueryExpr) Plan(tableInfo *models.TableInfo) (*models.QueryExecutionPlan, error) {
	return md.ast.calcQuery(tableInfo)
}

func (md *QueryExpr) String() string {
	return md.ast.String()
}
