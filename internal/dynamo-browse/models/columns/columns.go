package columns

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/queryexpr"
)

type Columns struct {
	TableInfo *models.TableInfo
	Columns   []Column
}

func NewColumnsFromResultSet(rs *models.ResultSet) *Columns {
	rsCols := rs.Columns()

	cols := make([]Column, len(rsCols))
	for i, c := range rsCols {
		cols[i] = Column{
			Name:      c,
			Evaluator: SimpleFieldValueEvaluator(c),
		}
	}

	return &Columns{
		TableInfo: rs.TableInfo,
		Columns:   cols,
	}
}

func (cols *Columns) VisibleColumns() []Column {
	if cols == nil {
		return []Column{}
	}

	visibleCols := make([]Column, 0)
	for _, col := range cols.Columns {
		if col.Hidden {
			continue
		}
		visibleCols = append(visibleCols, col)
	}
	return visibleCols
}

type Column struct {
	Name      string
	Evaluator FieldValueEvaluator
	Hidden    bool
}

type FieldValueEvaluator interface {
	EvaluateForItem(item models.Item) (types.AttributeValue, error)
}

type SimpleFieldValueEvaluator string

func (sfve SimpleFieldValueEvaluator) EvaluateForItem(item models.Item) (types.AttributeValue, error) {
	return item[string(sfve)], nil
}

type ExprFieldValueEvaluator struct {
	Expr *queryexpr.QueryExpr
}

func (sfve ExprFieldValueEvaluator) EvaluateForItem(item models.Item) (types.AttributeValue, error) {
	return sfve.Expr.EvalItem(item)
}
