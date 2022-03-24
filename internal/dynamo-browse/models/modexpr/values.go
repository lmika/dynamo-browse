package modexpr

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
	"strconv"
)

func (a *astLiteralValue) dynamoValue() (types.AttributeValue, error) {
	s, err := strconv.Unquote(a.String)
	if err != nil {
		return nil, errors.Wrap(err, "cannot unquote string")
	}
	return &types.AttributeValueMemberS{Value: s}, nil
}
