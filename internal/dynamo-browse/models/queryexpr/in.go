package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/common/sliceutils"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"strings"
)

func (a *astIn) evalToIR(info *models.TableInfo) (irAtom, error) {
	leftIR, err := a.Ref.evalToIR(info)
	if err != nil {
		return nil, err
	}

	if len(a.Operand) == 0 {
		return leftIR, nil
	}

	//singleName, isSingleName := a.Ref.leftOperandName()
	//if !isSingleName {
	//	return nil, errors.Errorf("%v: cannot use dereferences", singleName)
	//}

	nameIR, isNameIR := leftIR.(irNamePath)
	if !isNameIR {
		return nil, OperandNotANameError(a.Ref.String())
	}

	oprValues := make([]valueIRAtom, len(a.Operand))
	for i, o := range a.Operand {
		v, err := o.evalToIR(info)
		if err != nil {
			return nil, err
		}

		valueIR, isValueIR := v.(valueIRAtom)
		if !isValueIR {
			return nil, errors.Wrapf(ValueMustBeLiteralError{}, "'in' operand %v", i)
		}
		oprValues[i] = valueIR
	}

	// If there is a single operand value, and the name is either the partition or sort key, then
	// convert this to an equality so that it could be run as a query
	if len(oprValues) == 1 && (nameIR.keyName() == info.Keys.PartitionKey || nameIR.keyName() == info.Keys.SortKey) {
		return irFieldEq{name: nameIR, value: oprValues[0]}, nil
	}

	return irIn{name: nameIR, values: oprValues}, nil
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
	name   nameIRAtom
	values []valueIRAtom
}

//func (i irIn) operandFieldName() string {
//	return i.name
//}

func (i irIn) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	return false
}

func (i irIn) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	return expression.KeyConditionBuilder{}, errors.New("queries are not supported")
}

func (i irIn) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	right := expression.Value(i.values[0].goValue())
	others := sliceutils.Map(i.values[1:], func(x valueIRAtom) expression.OperandBuilder {
		return expression.Value(x.goValue())
	})

	return i.name.calcName(info).In(right, others...), nil
}
