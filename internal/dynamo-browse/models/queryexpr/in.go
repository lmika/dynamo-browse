package queryexpr

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
		case oprIRAtom:
			nameIR, isNameIR := leftIR.(irNamePath)
			if !isNameIR {
				return nil, OperandNotANameError(a.Ref.String())
			}

			ir = irIn{name: nameIR, values: []oprIRAtom{t}}
		case multiValueIRAtom:
			nameIR, isNameIR := leftIR.(irNamePath)
			if !isNameIR {
				return nil, OperandNotANameError(a.Ref.String())
			}

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

func (a *astIn) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
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
			cmp, isComparable := attrutils.CompareScalarAttributes(val, evalOp)
			if !isComparable {
				continue
			} else if cmp == 0 {
				return &types.AttributeValueMemberBOOL{Value: true}, nil
			}
		}
		return &types.AttributeValueMemberBOOL{Value: false}, nil
	case a.SingleOperand != nil:
		evalOp, err := a.SingleOperand.evalItem(ctx, item)
		if err != nil {
			return nil, err
		}

		switch t := evalOp.(type) {
		case *types.AttributeValueMemberS:
			str, canToStr := attrutils.AttributeToString(val)
			if !canToStr {
				return &types.AttributeValueMemberBOOL{Value: false}, nil
			}

			return &types.AttributeValueMemberBOOL{Value: strings.Contains(t.Value, str)}, nil
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
		case *types.AttributeValueMemberSS:
			str, canToStr := attrutils.AttributeToString(val)
			if !canToStr {
				return &types.AttributeValueMemberBOOL{Value: false}, nil
			}

			for _, listItem := range t.Value {
				if str != listItem {
					return &types.AttributeValueMemberBOOL{Value: false}, nil
				}
			}
			return &types.AttributeValueMemberBOOL{Value: true}, nil
		case *types.AttributeValueMemberBS:
			b, isB := val.(*types.AttributeValueMemberB)
			if !isB {
				return &types.AttributeValueMemberBOOL{Value: false}, nil
			}

			for _, listItem := range t.Value {
				if !bytes.Equal(b.Value, listItem) {
					return &types.AttributeValueMemberBOOL{Value: false}, nil
				}
			}
			return &types.AttributeValueMemberBOOL{Value: true}, nil
		case *types.AttributeValueMemberNS:
			n, isN := val.(*types.AttributeValueMemberN)
			if !isN {
				return &types.AttributeValueMemberBOOL{Value: false}, nil
			}

			for _, listItem := range t.Value {
				// TODO: this is not actually right
				if n.Value != listItem {
					return &types.AttributeValueMemberBOOL{Value: false}, nil
				}
			}
			return &types.AttributeValueMemberBOOL{Value: true}, nil
		case *types.AttributeValueMemberM:
			str, canToStr := attrutils.AttributeToString(val)
			if !canToStr {
				return &types.AttributeValueMemberBOOL{Value: false}, nil
			}
			_, hasItem := t.Value[str]
			return &types.AttributeValueMemberBOOL{Value: hasItem}, nil
		}
		return nil, ValuesNotInnableError{Val: evalOp}
	}
	return nil, errors.New("internal error: unhandled 'in' case")
}

func (a *astIn) canModifyItem(ctx *evalContext, item models.Item) bool {
	if len(a.Operand) != 0 || a.SingleOperand != nil {
		return false
	}
	return a.Ref.canModifyItem(ctx, item)
}

func (a *astIn) setEvalItem(ctx *evalContext, item models.Item, value types.AttributeValue) error {
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

type irContains struct {
	needle   valueIRAtom
	haystack nameIRAtom
}

func (i irContains) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	needle := i.needle.goValue()
	haystack := i.haystack.calcName(info)

	return haystack.Contains(fmt.Sprint(needle)), nil
}
