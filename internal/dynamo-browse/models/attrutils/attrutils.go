package attrutils

import (
	"math/big"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func CompareScalarAttributes(x, y types.AttributeValue) (int, bool) {
	switch xVal := x.(type) {
	case *types.AttributeValueMemberS:
		if yVal, ok := y.(*types.AttributeValueMemberS); ok {
			return comparisonValue(xVal.Value == yVal.Value, xVal.Value < yVal.Value), true
		}
	case *types.AttributeValueMemberN:
		if yVal, ok := y.(*types.AttributeValueMemberN); ok {
			xNumVal, _, err := big.ParseFloat(xVal.Value, 10, 63, big.ToNearestEven)
			if err != nil {
				return 0, false
			}

			yNumVal, _, err := big.ParseFloat(yVal.Value, 10, 63, big.ToNearestEven)
			if err != nil {
				return 0, false
			}

			return xNumVal.Cmp(yNumVal), true
		}
	case *types.AttributeValueMemberBOOL:
		if yVal, ok := y.(*types.AttributeValueMemberBOOL); ok {
			return comparisonValue(xVal.Value == yVal.Value, !xVal.Value), true
		}
	}
	return 0, false
}

func AttributeToString(x types.AttributeValue) (string, bool) {
	switch xVal := x.(type) {
	case *types.AttributeValueMemberS:
		return xVal.Value, true
	case *types.AttributeValueMemberN:
		return xVal.Value, true
	case *types.AttributeValueMemberBOOL:
		if xVal.Value {
			return "true", true
		} else {
			return "false", true
		}
	}
	return "", false
}

func comparisonValue(isEqual bool, isLess bool) int {
	if isEqual {
		return 0
	} else if isLess {
		return -1
	}
	return 1
}
