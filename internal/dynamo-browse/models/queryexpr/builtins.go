package queryexpr

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
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
}
