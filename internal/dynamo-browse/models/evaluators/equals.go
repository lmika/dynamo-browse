package evaluators

import (
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/queryexpr"
)

func Equals(x, y models.FieldValueEvaluator) bool {
	if x == nil {
		return y == nil
	}

	switch xt := x.(type) {
	case models.SimpleFieldValueEvaluator:
		if yt, ok := y.(models.SimpleFieldValueEvaluator); ok {
			return xt == yt
		}
	case queryexpr.ExprFieldValueEvaluator:
		if yt, ok := y.(queryexpr.ExprFieldValueEvaluator); ok {
			return xt.Expr.Equal(yt.Expr)
		}
	}

	return false
}
