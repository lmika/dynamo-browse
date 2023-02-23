package columns

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/queryexpr"
)

type Columns struct {
	TableInfo     *models.TableInfo
	WasRearranged bool
	Columns       []Column
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

func (cols *Columns) AddMissingColumns(rs *models.ResultSet) {
	existingColumns := make(map[string]Column)
	for _, col := range cols.Columns {
		existingColumns[col.Name] = col
	}

	rsCols := rs.Columns()
	var newCols []Column

	if cols.WasRearranged {
		newCols = append([]Column{}, cols.Columns...)
		for _, c := range rsCols {
			if _, hasCol := existingColumns[c]; !hasCol {
				newCols = append(newCols, Column{
					Name:      c,
					Evaluator: SimpleFieldValueEvaluator(c),
				})
			}
		}
	} else {
		newCols = make([]Column, len(rsCols))
		for i, c := range rsCols {
			if existingCol, hasCol := existingColumns[c]; hasCol {
				newCols[i] = existingCol
			} else {
				newCols[i] = Column{
					Name:      c,
					Evaluator: SimpleFieldValueEvaluator(c),
				}
			}
		}
	}

	cols.Columns = newCols
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
	EvaluateForItem(item models.Item) types.AttributeValue
}

type SimpleFieldValueEvaluator string

func (sfve SimpleFieldValueEvaluator) EvaluateForItem(item models.Item) types.AttributeValue {
	return item[string(sfve)]
}

type ExprFieldValueEvaluator struct {
	Expr *queryexpr.QueryExpr
}

func (sfve ExprFieldValueEvaluator) EvaluateForItem(item models.Item) types.AttributeValue {
	val, _ := sfve.Expr.EvalItem(item)
	return val
}
