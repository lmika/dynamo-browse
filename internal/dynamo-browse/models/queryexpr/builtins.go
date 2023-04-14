package queryexpr

import (
	"context"
	"github.com/pkg/errors"
)

type nativeFunc func(ctx context.Context, args []exprValue) (exprValue, error)

var nativeFuncs = map[string]nativeFunc{
	"size": func(ctx context.Context, args []exprValue) (exprValue, error) {
		if len(args) != 1 {
			return nil, InvalidArgumentNumberError{Name: "size", Expected: 1, Actual: len(args)}
		}

		var l int
		switch t := args[0].(type) {
		case stringExprValue:
			l = len(t)
		case mappableExprValue:
			l = t.len()
		case slicableExprValue:
			l = t.len()
		default:
			return nil, errors.New("cannot take size of arg")
		}
		return int64ExprValue(l), nil
	},

	"range": func(ctx context.Context, args []exprValue) (exprValue, error) {
		if len(args) != 2 {
			return nil, InvalidArgumentNumberError{Name: "range", Expected: 2, Actual: len(args)}
		}

		xVal, isXNum := args[0].(numberableExprValue)
		if !isXNum {
			return nil, InvalidArgumentTypeError{Name: "range", ArgIndex: 0, Expected: "N"}
		}
		yVal, isYNum := args[1].(numberableExprValue)
		if !isYNum {
			return nil, InvalidArgumentTypeError{Name: "range", ArgIndex: 1, Expected: "N"}
		}

		xInt, _ := xVal.asBigFloat().Int64()
		yInt, _ := yVal.asBigFloat().Int64()
		xs := make([]exprValue, 0)
		for x := xInt; x <= yInt; x++ {
			xs = append(xs, int64ExprValue(x))
		}
		return listExprValue(xs), nil
	},

	"_x_now": func(ctx context.Context, args []exprValue) (exprValue, error) {
		now := timeSourceFromContext(ctx).now().Unix()
		return int64ExprValue(now), nil
	},

	"_x_add": func(ctx context.Context, args []exprValue) (exprValue, error) {
		if len(args) != 2 {
			return nil, InvalidArgumentNumberError{Name: "_x_add", Expected: 2, Actual: len(args)}
		}

		xVal, isXNum := args[0].(numberableExprValue)
		if !isXNum {
			return nil, InvalidArgumentTypeError{Name: "_x_add", ArgIndex: 0, Expected: "N"}
		}
		yVal, isYNum := args[1].(numberableExprValue)
		if !isYNum {
			return nil, InvalidArgumentTypeError{Name: "_x_add", ArgIndex: 1, Expected: "N"}
		}

		return bigNumExprValue{num: xVal.asBigFloat().Add(xVal.asBigFloat(), yVal.asBigFloat())}, nil
	},

	"_x_concat": func(ctx context.Context, args []exprValue) (exprValue, error) {
		if len(args) != 2 {
			return nil, InvalidArgumentNumberError{Name: "_x_concat", Expected: 2, Actual: len(args)}
		}

		xVal, isXNum := args[0].(stringableExprValue)
		if !isXNum {
			return nil, InvalidArgumentTypeError{Name: "_x_concat", ArgIndex: 0, Expected: "S"}
		}
		yVal, isYNum := args[1].(stringableExprValue)
		if !isYNum {
			return nil, InvalidArgumentTypeError{Name: "_x_concat", ArgIndex: 1, Expected: "S"}
		}

		return stringExprValue(xVal.asString() + yVal.asString()), nil
	},
}
