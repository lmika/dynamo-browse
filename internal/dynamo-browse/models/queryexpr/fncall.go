package queryexpr

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/common/sliceutils"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"strings"
)

func (a *astFunctionCall) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	callerIr, err := a.Caller.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}
	if !a.IsCall {
		return callerIr, nil
	}

	nameIr, isNameIr := callerIr.(nameIRAtom)
	if !isNameIr || nameIr.keyName() == "" {
		return nil, OperandNotANameError("")
	}

	irNodes, err := sliceutils.MapWithError(a.Args, func(x *astExpr) (irAtom, error) { return x.evalToIR(ctx, info) })
	if err != nil {
		return nil, err
	}

	// TODO: do this properly
	switch nameIr.keyName() {
	case "size":
		if len(irNodes) != 1 {
			return nil, InvalidArgumentNumberError{Name: "size", Expected: 1, Actual: len(irNodes)}
		}
		name, isName := irNodes[0].(nameIRAtom)
		if !isName {
			return nil, OperandNotANameError(a.Args[0].String())
		}
		return irSizeFn{name}, nil
	case "range":
		if len(irNodes) != 2 {
			return nil, InvalidArgumentNumberError{Name: "range", Expected: 2, Actual: len(irNodes)}
		}

		// TEMP
		fromVal := irNodes[0].(valueIRAtom).goValue().(int64)
		toVal := irNodes[1].(valueIRAtom).goValue().(int64)
		return irRangeFn{fromVal, toVal}, nil
	}
	return nil, UnrecognisedFunctionError{Name: nameIr.keyName()}
}

func (a *astFunctionCall) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
	if !a.IsCall {
		return a.Caller.evalItem(ctx, item)
	}

	name, isName := a.Caller.unqualifiedName()
	if !isName {
		return nil, OperandNotANameError(a.Args[0].String())
	}
	fn, isFn := nativeFuncs[name]
	if !isFn {
		return nil, UnrecognisedFunctionError{Name: name}
	}

	args, err := sliceutils.MapWithError(a.Args, func(a *astExpr) (types.AttributeValue, error) {
		return a.evalItem(ctx, item)
	})
	if err != nil {
		return nil, err
	}

	return fn(context.Background(), args)
}

func (a *astFunctionCall) canModifyItem(ctx *evalContext, item models.Item) bool {
	// TODO: Should a function vall return an item?
	if a.IsCall {
		return false
	}
	return a.Caller.canModifyItem(ctx, item)
}

func (a *astFunctionCall) setEvalItem(ctx *evalContext, item models.Item, value types.AttributeValue) error {
	// TODO: Should a function vall return an item?
	if a.IsCall {
		return PathNotSettableError{}
	}
	return a.Caller.setEvalItem(ctx, item, value)
}

func (a *astFunctionCall) deleteAttribute(ctx *evalContext, item models.Item) error {
	// TODO: Should a function vall return an item?
	if a.IsCall {
		return PathNotSettableError{}
	}
	return a.Caller.deleteAttribute(ctx, item)
}

func (a *astFunctionCall) String() string {
	var sb strings.Builder

	sb.WriteString(a.Caller.String())
	if a.IsCall {
		sb.WriteRune('(')
		for i, q := range a.Args {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(q.String())
		}
		sb.WriteRune(')')
	}
	return sb.String()
}

type irSizeFn struct {
	arg nameIRAtom
}

func (i irSizeFn) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	return expression.ConditionBuilder{}, errors.New("cannot run as scan")
}

func (i irSizeFn) calcOperand(info *models.TableInfo) expression.OperandBuilder {
	name := i.arg.calcName(info)
	return name.Size()
}

type irRangeFn struct {
	fromIdx int64
	toIdx   int64
}

func (i irRangeFn) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	return expression.ConditionBuilder{}, errors.New("cannot run as scan")
}

func (i irRangeFn) calcGoValues(info *models.TableInfo) ([]any, error) {
	xs := make([]any, 0)
	for x := i.fromIdx; x <= i.toIdx; x++ {
		xs = append(xs, x)
	}
	return xs, nil
}
