package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"strconv"

	"github.com/pkg/errors"
)

func (a *astLiteralValue) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	v, err := a.exprValue()
	if err != nil {
		return nil, err
	}
	return irValue{value: v}, nil
}

func (a *astLiteralValue) exprValue() (exprValue, error) {
	switch {
	case a.StringVal != nil:
		s, err := strconv.Unquote(*a.StringVal)
		if err != nil {
			return nil, errors.Wrap(err, "cannot unquote string")
		}
		return stringExprValue(s), nil
	case a.IntVal != nil:
		return int64ExprValue(*a.IntVal), nil
	case a.TrueBoolValue:
		return boolExprValue(true), nil
	case a.FalseBoolValue:
		return boolExprValue(false), nil
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
	case a.TrueBoolValue:
		return "true"
	case a.FalseBoolValue:
		return "false"
	}
	return ""
}

type irValue struct {
	value exprValue
}

func (i irValue) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	return expression.ConditionBuilder{}, NodeCannotBeConvertedToQueryError{}
}

func (i irValue) exprValue() exprValue {
	return i.value
}

func (a irValue) calcOperand(info *models.TableInfo) expression.OperandBuilder {
	return expression.Value(a.value.asGoValue())
}
