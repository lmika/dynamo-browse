package queryexpr

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
)

func (a *astBetweenOp) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	leftIR, err := a.Ref.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}

	if a.From == nil {
		return leftIR, nil
	}

	nameIR, isNameIR := leftIR.(nameIRAtom)
	if !isNameIR {
		return nil, OperandNotANameError(a.Ref.String())
	}

	fromIR, err := a.From.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}
	toIR, err := a.To.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}

	fromOprIR, isFromOprIR := fromIR.(oprIRAtom)
	if !isFromOprIR {
		return nil, OperandNotAnOperandError{}
	}
	toOprIR, isToOprIR := toIR.(oprIRAtom)
	if !isToOprIR {
		return nil, OperandNotAnOperandError{}
	}

	return irBetween{name: nameIR, from: fromOprIR, to: toOprIR}, nil
}

func (a *astBetweenOp) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
	val, err := a.Ref.evalItem(ctx, item)
	if a.From == nil {
		return val, err
	}

	panic("TODO")
}

func (a *astBetweenOp) canModifyItem(ctx *evalContext, item models.Item) bool {
	if a.From != nil {
		return false
	}
	return a.Ref.canModifyItem(ctx, item)
}

func (a *astBetweenOp) setEvalItem(ctx *evalContext, item models.Item, value types.AttributeValue) error {
	if a.From != nil {
		return PathNotSettableError{}
	}
	return a.Ref.setEvalItem(ctx, item, value)
}

func (a *astBetweenOp) deleteAttribute(ctx *evalContext, item models.Item) error {
	if a.From != nil {
		return PathNotSettableError{}
	}
	return a.Ref.deleteAttribute(ctx, item)
}

func (a *astBetweenOp) String() string {
	name := a.Ref.String()
	if a.From != nil {
		return fmt.Sprintf("%v between %v and %v", name, a.From.String(), a.To.String())
	}
	return name
}

type irBetween struct {
	name nameIRAtom
	from oprIRAtom
	to   oprIRAtom
}

func (i irBetween) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	nb := i.name.calcName(info)
	fb := i.from.calcOperand(info)
	tb := i.to.calcOperand(info)

	return nb.Between(fb, tb), nil
}
