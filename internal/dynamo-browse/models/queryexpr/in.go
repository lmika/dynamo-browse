package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/common/sliceutils"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"strings"
)

func (a *astIn) evalToIR(info *models.TableInfo) (irAtom, error) {
	if len(a.Operand) == 0 {
		return a.Ref.evalToIR(info)
	}

	singleName, isSingleName := a.Ref.leftOperandName()
	if !isSingleName {
		return nil, errors.Errorf("%v: cannot use dereferences", singleName)
	}

	oprValues := make([]any, len(a.Operand))
	for i, o := range a.Operand {
		v, err := o.rightOperandGoValue()
		if err != nil {
			return nil, errors.Errorf("'in' operand %v: %v", i, o)
		}
		oprValues[i] = v
	}

	// If there is a single operand value, and the name is either the partition or sort key, then
	// convert this to an equality so that it could be run as a query
	if len(oprValues) == 1 && (singleName == info.Keys.PartitionKey || singleName == info.Keys.SortKey) {
		return irFieldEq{name: singleName, value: oprValues[0]}, nil
	}

	return irIn{name: singleName, values: oprValues}, nil
}

func (a *astIn) String() string {
	var sb strings.Builder

	sb.WriteString(a.Ref.String())
	sb.WriteString(" in (")
	for i, o := range a.Operand {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(o.String())
	}
	sb.WriteString(")")

	return sb.String()
}

type irIn struct {
	name   string
	values []any
}

func (i irIn) operandFieldName() string {
	return i.name
}

func (i irIn) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	return false
}

func (i irIn) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	return expression.KeyConditionBuilder{}, errors.New("queries are not supported")
}

func (i irIn) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	right := expression.Value(i.values[0])
	others := sliceutils.Map(i.values[1:], func(x any) expression.OperandBuilder {
		return expression.Value(x)
	})

	return expression.Name(i.name).In(right, others...), nil
}
