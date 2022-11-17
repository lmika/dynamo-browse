package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/common/sliceutils"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/attrutils"
	"github.com/pkg/errors"
	"strings"
)

func (a *astIn) evalToIR(info *models.TableInfo) (irAtom, error) {
	leftIR, err := a.Ref.evalToIR(info)
	if err != nil {
		return nil, err
	}

	if len(a.Operand) == 0 && a.SingleOperand == nil {
		return leftIR, nil
	}

	nameIR, isNameIR := leftIR.(irNamePath)
	if !isNameIR {
		return nil, OperandNotANameError(a.Ref.String())
	}

	var ir irAtom
	switch {
	case len(a.Operand) > 0:
		oprValues := make([]oprIRAtom, len(a.Operand))
		for i, o := range a.Operand {
			v, err := o.evalToIR(info)
			if err != nil {
				return nil, err
			}

			valueIR, isValueIR := v.(oprIRAtom)
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

			if valueIR, isValueIR := oprValues[0].(valueIRAtom); isValueIR {
				return irKeyFieldEq{name: nameIR, value: valueIR}, nil
			}
			return irGenericEq{name: nameIR, value: oprValues[0]}, nil
		}

		ir = irIn{name: nameIR, values: oprValues}
	case a.SingleOperand != nil:
		oprs, err := a.SingleOperand.evalToIR(info)
		if err != nil {
			return nil, err
		}

		switch t := oprs.(type) {
		case oprIRAtom:
			ir = irIn{name: nameIR, values: []oprIRAtom{t}}
		case multiValueIRAtom:
			ir = irLiteralValues{name: nameIR, values: t}
		default:
			return nil, OperandNotAnOperandError{}
		}
	}

	if a.HasNot {
		return &irBoolNot{atom: ir}, nil
	}
	return ir, nil
}

func (a *astIn) evalItem(item models.Item) (types.AttributeValue, error) {
	val, err := a.Ref.evalItem(item)
	if err != nil {
		return nil, err
	}
	if len(a.Operand) == 0 && a.SingleOperand == nil {
		return val, nil
	}

	switch {
	case len(a.Operand) > 0:
		for _, opr := range a.Operand {
			evalOp, err := opr.evalItem(item)
			if err != nil {
				return nil, err
			}
			cmp, isComparable := attrutils.CompareScalarAttributes(val, evalOp)
			if !isComparable {
				continue
			} else if cmp == 0 {
				return &types.AttributeValueMemberBOOL{Value: true}, nil
			}
		}
		return &types.AttributeValueMemberBOOL{Value: false}, nil
	case a.SingleOperand != nil:
		evalOp, err := a.SingleOperand.evalItem(item)
		if err != nil {
			return nil, err
		}

		switch t := evalOp.(type) {
		case *types.AttributeValueMemberL:
			for _, listItem := range t.Value {
				cmp, isComparable := attrutils.CompareScalarAttributes(val, listItem)
				if !isComparable {
					continue
				} else if cmp == 0 {
					return &types.AttributeValueMemberBOOL{Value: true}, nil
				}
			}
			return &types.AttributeValueMemberBOOL{Value: false}, nil
		case *types.AttributeValueMemberM:
			str, canToStr := attrutils.AttributeToString(val)
			if !canToStr {
				return &types.AttributeValueMemberBOOL{Value: false}, nil
			}
			_, hasItem := t.Value[str]
			return &types.AttributeValueMemberBOOL{Value: hasItem}, nil
			// TODO: the sets
		}
		return nil, ValuesNotInnableError{Val: evalOp}
	}
	return nil, errors.New("internal error: unhandled 'in' case")
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
	values []oprIRAtom
}

func (i irIn) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	right := i.values[0].calcOperand(info)
	others := sliceutils.Map(i.values[1:], func(x oprIRAtom) expression.OperandBuilder {
		return x.calcOperand(info)
	})

	return i.name.calcName(info).In(right, others...), nil
}

type irLiteralValues struct {
	name   nameIRAtom
	values multiValueIRAtom
}

func (i irLiteralValues) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	vals, err := i.values.calcGoValues(info)
	if err != nil {
		return expression.ConditionBuilder{}, err
	}

	oprValues := sliceutils.Map(vals, func(t any) expression.OperandBuilder {
		return expression.Value(t)
	})
	return i.name.calcName(info).In(oprValues[0], oprValues[1:]...), nil
}
