package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/common/sliceutils"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"strings"
)

func (a *astFunctionCall) evalToIR(info *models.TableInfo) (irAtom, error) {
	callerIr, err := a.Caller.evalToIR(info)
	if err != nil {
		return nil, err
	}
	if !a.IsCall {
		return callerIr, nil
	}

	nameIr, isNameIr := callerIr.(nameIRAtom)
	if !isNameIr || nameIr.keyName() == "" {
		return nil, OperandNotANameError("")
	}

	irNodes, err := sliceutils.MapWithError(a.Args, func(x *astExpr) (irAtom, error) { return x.evalToIR(info) })
	if err != nil {
		return nil, err
	}

	// TODO: do this properly
	switch nameIr.keyName() {
	case "size":
		if len(irNodes) != 1 {
			return nil, InvalidArgumentNumberError{Name: "size", Expected: 1, Actual: len(irNodes)}
		}
		name, isName := irNodes[0].(nameIRAtom)
		if !isName {
			return nil, OperandNotANameError(a.Args[0].String())
		}
		return irSizeFn{name}, nil
	}
	return nil, UnrecognisedFunctionError{Name: nameIr.keyName()}
}

func (a *astFunctionCall) evalItem(item models.Item) (types.AttributeValue, error) {
	if !a.IsCall {
		return a.Caller.evalItem(item)
	}
	panic("TODO")
}

func (a *astFunctionCall) String() string {
	var sb strings.Builder

	sb.WriteString(a.Caller.String())
	if a.IsCall {
		sb.WriteRune('(')
		for i, q := range a.Args {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(q.String())
		}
		sb.WriteRune(')')
	}
	return sb.String()
}

type irSizeFn struct {
	arg nameIRAtom
}

func (i irSizeFn) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	return expression.ConditionBuilder{}, errors.New("cannot run as scan")
}

func (i irSizeFn) calcOperand(info *models.TableInfo) expression.OperandBuilder {
	name := i.arg.calcName(info)
	return name.Size()
}
