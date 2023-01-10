package attrutils

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func Equals(x, y types.AttributeValue) bool {
	switch xVal := x.(type) {
	case *types.AttributeValueMemberS:
		c, ok := CompareScalarAttributes(x, y)
		return ok && c == 0
	case *types.AttributeValueMemberN:
		c, ok := CompareScalarAttributes(x, y)
		return ok && c == 0
	case *types.AttributeValueMemberBOOL:
		c, ok := CompareScalarAttributes(x, y)
		return ok && c == 0
	case *types.AttributeValueMemberB:
		if yVal, ok := y.(*types.AttributeValueMemberB); ok {
			return slices.Equal(xVal.Value, yVal.Value)
		}
	case *types.AttributeValueMemberNULL:
		if yVal, ok := y.(*types.AttributeValueMemberNULL); ok {
			return xVal.Value == yVal.Value
		}
	case *types.AttributeValueMemberL:
		if yVal, ok := y.(*types.AttributeValueMemberL); ok {
			return slices.EqualFunc(xVal.Value, yVal.Value, Equals)
		}
	case *types.AttributeValueMemberM:
		if yVal, ok := y.(*types.AttributeValueMemberM); ok {
			return maps.EqualFunc(xVal.Value, yVal.Value, Equals)
		}
	case *types.AttributeValueMemberBS:
		if yVal, ok := y.(*types.AttributeValueMemberBS); ok {
			return slices.EqualFunc(xVal.Value, yVal.Value, func(xs, ys []byte) bool {
				return slices.Equal(xs, ys)
			})
		}
	case *types.AttributeValueMemberNS:
		if yVal, ok := y.(*types.AttributeValueMemberNS); ok {
			return slices.Equal(xVal.Value, yVal.Value)
		}
	case *types.AttributeValueMemberSS:
		if yVal, ok := y.(*types.AttributeValueMemberSS); ok {
			return slices.Equal(xVal.Value, yVal.Value)
		}
	}

	return false
}
