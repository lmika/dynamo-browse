package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
)

func (a *astLiteralValue) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	v, err := a.goValue()
	if err != nil {
		return nil, err
	}
	return irValue{value: v}, nil
}

func (a *astLiteralValue) dynamoValue() (types.AttributeValue, error) {
	if a == nil {
		return nil, nil
	}

	goValue, err := a.goValue()
	if err != nil {
		return nil, err
	}

	switch v := goValue.(type) {
	case string:
		return &types.AttributeValueMemberS{Value: v}, nil
	case int64:
		return &types.AttributeValueMemberN{Value: strconv.FormatInt(v, 10)}, nil
	}

	return nil, errors.New("unrecognised type")
}

func (a *astLiteralValue) goValue() (any, error) {
	if a == nil {
		return nil, nil
	}

	switch {
	case a.StringVal != nil:
		s, err := strconv.Unquote(*a.StringVal)
		if err != nil {
			return nil, errors.Wrap(err, "cannot unquote string")
		}
		return s, nil
	case a.IntVal != nil:
		return *a.IntVal, nil
	}
	return nil, errors.New("unrecognised type")
}

func (a *astLiteralValue) String() string {
	if a == nil {
		return ""
	}

	switch {
	case a.StringVal != nil:
		return *a.StringVal
	case a.IntVal != nil:
		return strconv.FormatInt(*a.IntVal, 10)
	}
	return ""
}

type irValue struct {
	value any
}

func (i irValue) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	return expression.ConditionBuilder{}, NodeCannotBeConvertedToQueryError{}
}

func (i irValue) goValue() any {
	return i.value
}

func (a irValue) calcOperand(info *models.TableInfo) expression.OperandBuilder {
	return expression.Value(a.goValue())
}
