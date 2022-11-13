package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
)

// TO DELETE = operandFieldName() string

type irAtom interface {
	// calcQueryForScan returns the condition builder for this atom to include in a scan
	calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error)
}

type queryableIRAtom interface {
	irAtom

	// canBeExecutedAsQuery returns true if the atom is capable of being executed as a query
	canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool

	// calcQueryForQuery returns a key condition builder for this atom to include in a query
	calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error)
}

type oprIRAtom interface {
	calcOperand(info *models.TableInfo) expression.OperandBuilder
}

type nameIRAtom interface {
	oprIRAtom

	// keyName returns the name as key if it can be a DB key.  Returns "" if this name cannot be a key
	keyName() string
	calcName(info *models.TableInfo) expression.NameBuilder
}

type valueIRAtom interface {
	oprIRAtom
	goValue() any
}

func canExecuteAsQuery(ir irAtom, info *models.TableInfo, qci *queryCalcInfo) bool {
	queryable, isQuearyable := ir.(queryableIRAtom)
	if !isQuearyable {
		return false
	}
	return queryable.canBeExecutedAsQuery(info, qci)
}
