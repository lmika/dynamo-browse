package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
)

type ExprFieldValueEvaluator struct {
	Expr *QueryExpr
}

func (sfve ExprFieldValueEvaluator) EvaluateForItem(item models.Item) types.AttributeValue {
	val, _ := sfve.Expr.EvalItem(item)
	return val
}
