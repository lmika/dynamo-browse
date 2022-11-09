package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
)

type expressionBuilder any

type irAtom interface {
	// operandFieldName returns the field that this atom operates on.  For example,
	// if this IR node represents 'a = "b"', this should return "a".
	// If this does not operate on a definitive field name, this returns null
	//operandFieldName() string

	// canBeExecutedAsQuery returns true if the atom is capable of being executed as a query
	canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool

	// calcQueryForQuery returns a key condition builder for this atom to include in a query
	calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error)

	// calcQueryForScan returns the condition builder for this atom to include in a scan
	calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error)
}

type nameIRAtom interface {
	// keyName returns the name as key if it can be a DB key.  Returns "" if this name cannot be a key
	keyName() string
	calcName(info *models.TableInfo) expression.NameBuilder
}

type valueIRAtom interface {
	goValue() any
}

//type scanQueryCalc interface {
//}
//
//type conditionScanQueryCalc struct {
//	expression.ConditionBuilder
//}
//
//type nameScanQueryCalc struct {
//	expression.NameBuilder
//}
//
//type valueScanQueryCalc struct {
//	value any
//}
