package queryexpr

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
)

func (a *astLiteralValue) dynamoValue() (types.AttributeValue, error) {
	s, err := strconv.Unquote(a.StringVal)
	if err != nil {
		return nil, errors.Wrap(err, "cannot unquote string")
	}
	return &types.AttributeValueMemberS{Value: s}, nil
}

func (a *astLiteralValue) goValue() (any, error) {
	s, err := strconv.Unquote(a.StringVal)
	if err != nil {
		return nil, errors.Wrap(err, "cannot unquote string")
	}
	return s, nil
}
