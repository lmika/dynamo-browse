package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
)

//type irDisjunction struct {
//	conj []*irConjunction
//}

//type irConjunction struct {
//	atoms []irAtom
//}

type irAtom interface {
	// operandFieldName returns the field that this atom operates on.  For example,
	// if this IR node represents 'a = "b"', this should return "a".
	// If this does not operate on a definitive field name, this returns null
	operandFieldName() string

	// canBeExecutedAsQuery returns true if the atom is capable of being executed as a query
	canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool

	// calcQueryForQuery returns a key condition builder for this atom to include in a query
	calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error)

	// calcQueryForScan returns the condition builder for this atom to include in a scan
	calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error)
}

//type irFieldEq struct {
//	name  string
//	value any
//}
//
//type irFieldBeginsWith struct {
//	name   string
//	prefix string
//}
