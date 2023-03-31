package queryexpr

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
)

type nativeFunc func(ctx context.Context, args []types.AttributeValue) (types.AttributeValue, error)

var nativeFuncs = map[string]nativeFunc{
	"size": func(ctx context.Context, args []types.AttributeValue) (types.AttributeValue, error) {
		if len(args) != 1 {
			return nil, InvalidArgumentNumberError{Name: "size", Expected: 1, Actual: len(args)}
		}

		var l int
		switch t := args[0].(type) {
		case *types.AttributeValueMemberB:
			l = len(t.Value)
		case *types.AttributeValueMemberS:
			l = len(t.Value)
		case *types.AttributeValueMemberL:
			l = len(t.Value)
		case *types.AttributeValueMemberM:
			l = len(t.Value)
		case *types.AttributeValueMemberSS:
			l = len(t.Value)
		case *types.AttributeValueMemberNS:
			l = len(t.Value)
		case *types.AttributeValueMemberBS:
			l = len(t.Value)
		default:
			return nil, errors.New("cannot take size of arg")
		}
		return &types.AttributeValueMemberN{Value: strconv.Itoa(l)}, nil
	},

	"_x_now": func(ctx context.Context, args []types.AttributeValue) (types.AttributeValue, error) {
		now := timeSourceFromContext(ctx).now().Unix()
		return &types.AttributeValueMemberN{Value: strconv.FormatInt(now, 10)}, nil
	},

	"_x_add": func(ctx context.Context, args []types.AttributeValue) (types.AttributeValue, error) {
		if len(args) != 2 {
			return nil, InvalidArgumentNumberError{Name: "_x_add", Expected: 2, Actual: len(args)}
		}

		xVal, isXNum := args[0].(*types.AttributeValueMemberN)
		if !isXNum {
			return nil, InvalidArgumentTypeError{Name: "_x_add", ArgIndex: 0, Expected: "N"}
		}
		yVal, isYNum := args[1].(*types.AttributeValueMemberN)
		if !isYNum {
			return nil, InvalidArgumentTypeError{Name: "_x_add", ArgIndex: 1, Expected: "N"}
		}

		xNumVal, _, err := big.ParseFloat(xVal.Value, 10, 63, big.ToNearestEven)
		if err != nil {
			return nil, err
		}

		yNumVal, _, err := big.ParseFloat(yVal.Value, 10, 63, big.ToNearestEven)
		if err != nil {
			return nil, err
		}

		return &types.AttributeValueMemberN{Value: xNumVal.Add(xNumVal, yNumVal).String()}, nil
	},
}
