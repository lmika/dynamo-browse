package attrutils

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func Truthy(x types.AttributeValue) bool {
	switch xVal := x.(type) {
	case *types.AttributeValueMemberS:
		return len(xVal.Value) > 0
	case *types.AttributeValueMemberN:
		return len(xVal.Value) > 0 && xVal.Value != "0"
	case *types.AttributeValueMemberBOOL:
		return xVal.Value
	case *types.AttributeValueMemberB:
		return len(xVal.Value) > 0
	case *types.AttributeValueMemberNULL:
		return !xVal.Value
	case *types.AttributeValueMemberL:
		return len(xVal.Value) > 0
	case *types.AttributeValueMemberM:
		return len(xVal.Value) > 0
	case *types.AttributeValueMemberBS:
		return len(xVal.Value) > 0
	case *types.AttributeValueMemberNS:
		return len(xVal.Value) > 0
	case *types.AttributeValueMemberSS:
		return len(xVal.Value) > 0
	}

	return false
}
