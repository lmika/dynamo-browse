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
		if a.HasNot {
			return irFieldNe{name: nameIR, value: oprValues[0]}, nil
		}
		return irKeyFieldEq{name: nameIR, value: oprValues[0]}, nil
	}

	var ir = irIn{name: nameIR, values: oprValues}
	if a.HasNot {
		return &irBoolNot{atom: ir}, nil
	}
	return ir, nil
}

func (a *astIn) String() string {
	var sb strings.Builder

	sb.WriteString(a.Ref.String())
	if a.HasNot {
		sb.WriteString(" not in (")
	} else {
		sb.WriteString(" in (")
	}

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

func (i irIn) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	right := expression.Value(i.values[0].goValue())
	others := sliceutils.Map(i.values[1:], func(x valueIRAtom) expression.OperandBuilder {
		return expression.Value(x.goValue())
	})

	return i.name.calcName(info).In(right, others...), nil
}
