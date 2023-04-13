package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/common/sliceutils"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/attrutils"
	"github.com/pkg/errors"
	"strings"
)

func (a *astIn) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	leftIR, err := a.Ref.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}

	if len(a.Operand) == 0 && a.SingleOperand == nil {
		return leftIR, nil
	}

	var ir irAtom
	switch {
	case len(a.Operand) > 0:
		nameIR, isNameIR := leftIR.(irNamePath)
		if !isNameIR {
			return nil, OperandNotANameError(a.Ref.String())
		}

		oprValues := make([]oprIRAtom, len(a.Operand))
		for i, o := range a.Operand {
			v, err := o.evalToIR(ctx, info)
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
		oprs, err := a.SingleOperand.evalToIR(ctx, info)
		if err != nil {
			return nil, err
		}

		switch t := oprs.(type) {
		case irNamePath:
			lit, isLit := leftIR.(valueIRAtom)
			if !isLit {
				return nil, OperandNotANameError(a.Ref.String())
			}
			ir = irContains{needle: lit, haystack: t}
		case valueIRAtom:
			nameIR, isNameIR := leftIR.(irNamePath)
			if !isNameIR {
				return nil, OperandNotANameError(a.Ref.String())
			}

			ir = irLiteralValues{name: nameIR, values: t}
		case oprIRAtom:
			nameIR, isNameIR := leftIR.(irNamePath)
			if !isNameIR {
				return nil, OperandNotANameError(a.Ref.String())
			}

			ir = irIn{name: nameIR, values: []oprIRAtom{t}}
		default:
			return nil, OperandNotAnOperandError{}
		}
	}

	if a.HasNot {
		return &irBoolNot{atom: ir}, nil
	}
	return ir, nil
}

func (a *astIn) evalItem(ctx *evalContext, item models.Item) (exprValue, error) {
	val, err := a.Ref.evalItem(ctx, item)
	if err != nil {
		return nil, err
	}
	if len(a.Operand) == 0 && a.SingleOperand == nil {
		return val, nil
	}

	switch {
	case len(a.Operand) > 0:
		for _, opr := range a.Operand {
			evalOp, err := opr.evalItem(ctx, item)
			if err != nil {
				return nil, err
			}
			// TODO: use native types here
			cmp, isComparable := attrutils.CompareScalarAttributes(val.asAttributeValue(), evalOp.asAttributeValue())
			if !isComparable {
				continue
			} else if cmp == 0 {
				return boolExprValue(true), nil
			}
		}
		return boolExprValue(false), nil
	case a.SingleOperand != nil:
		evalOp, err := a.SingleOperand.evalItem(ctx, item)
		if err != nil {
			return nil, err
		}

		switch t := evalOp.(type) {
		case stringableExprValue:
			str, canToStr := val.(stringableExprValue)
			if !canToStr {
				return boolExprValue(false), nil
			}

			return boolExprValue(strings.Contains(t.asString(), str.asString())), nil
		case slicableExprValue:
			for i := 0; i < t.len(); i++ {
				va, err := t.valueAt(i)
				if err != nil {
					return nil, err
				}

				// TODO: use expr value types here
				cmp, isComparable := attrutils.CompareScalarAttributes(val.asAttributeValue(), va.asAttributeValue())
				if !isComparable {
					continue
				} else if cmp == 0 {
					return boolExprValue(true), nil
				}
			}
			return boolExprValue(false), nil
		case mappableExprValue:
			str, canToStr := val.(stringableExprValue)
			if !canToStr {
				return boolExprValue(false), nil
			}
			hasKey := t.hasKey(str.asString())
			return boolExprValue(hasKey), nil
		}
		return nil, ValuesNotInnableError{Val: evalOp.asAttributeValue()}
	}
	return nil, errors.New("internal error: unhandled 'in' case")
}

func (a *astIn) canModifyItem(ctx *evalContext, item models.Item) bool {
	if len(a.Operand) != 0 || a.SingleOperand != nil {
		return false
	}
	return a.Ref.canModifyItem(ctx, item)
}

func (a *astIn) setEvalItem(ctx *evalContext, item models.Item, value exprValue) error {
	if len(a.Operand) != 0 || a.SingleOperand != nil {
		return PathNotSettableError{}
	}
	return a.Ref.setEvalItem(ctx, item, value)
}

func (a *astIn) deleteAttribute(ctx *evalContext, item models.Item) error {
	if len(a.Operand) != 0 || a.SingleOperand != nil {
		return PathNotSettableError{}
	}
	return a.Ref.deleteAttribute(ctx, item)

}

func (a *astIn) String() string {
	if len(a.Operand) == 0 && a.SingleOperand == nil {
		return a.Ref.String()
	}

	var sb strings.Builder

	sb.WriteString(a.Ref.String())
	if a.HasNot {
		sb.WriteString(" not in ")
	} else {
		sb.WriteString(" in ")
	}

	switch {
	case len(a.Operand) > 0:
		sb.WriteString("(")
		for i, o := range a.Operand {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(o.String())
		}
		sb.WriteString(")")
	case a.SingleOperand != nil:
		sb.WriteString(a.SingleOperand.String())
	}

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
	values valueIRAtom
}

func (iv irLiteralValues) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	if sliceable, isSliceable := iv.values.exprValue().(slicableExprValue); isSliceable {
		if sliceable.len() == 1 {
			va, err := sliceable.valueAt(0)
			if err != nil {
				return expression.ConditionBuilder{}, err
			}

			return iv.name.calcName(info).In(buildExpressionFromValue(va)), nil
		} else if sliceable.len() == 0 {
			// name is not in an empty slice, so this branch always evaluates to false
			// TODO: would be better to not even include this branch in some way?
			return expression.Equal(expression.Value(false), expression.Value(true)), nil
		}

		items := make([]expression.OperandBuilder, sliceable.len())
		for i := 0; i < sliceable.len(); i++ {
			va, err := sliceable.valueAt(i)
			if err != nil {
				return expression.ConditionBuilder{}, err
			}

			items[i] = buildExpressionFromValue(va)
		}

		return iv.name.calcName(info).In(items[0], items[1:]...), nil
	}

	return iv.name.calcName(info).In(buildExpressionFromValue(iv.values.exprValue())), nil
}

type irContains struct {
	needle   valueIRAtom
	haystack nameIRAtom
}

func (i irContains) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	strNeedle, isString := i.needle.exprValue().(stringableExprValue)
	if !isString {
		return expression.ConditionBuilder{}, errors.New("value cannot be converted to string")
	}

	haystack := i.haystack.calcName(info)
	return haystack.Contains(strNeedle.asString()), nil
}
