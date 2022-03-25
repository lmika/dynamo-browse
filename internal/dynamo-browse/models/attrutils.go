package models

import (
	"math/big"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func compareScalarAttributes(x, y types.AttributeValue) (int, bool) {
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

func comparisonValue(isEqual bool, isLess bool) int {
	if isEqual {
		return 0
	} else if isLess {
		return -1
	}
	return 1
}
