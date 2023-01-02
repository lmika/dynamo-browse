package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"strings"
)

func (a *astBooleanNot) evalToIR(ctx *evalContext, tableInfo *models.TableInfo) (irAtom, error) {
	irNode, err := a.Operand.evalToIR(ctx, tableInfo)
	if err != nil {
		return nil, err
	}

	if !a.HasNot {
		return irNode, nil
	}

	return &irBoolNot{atom: irNode}, nil
}

func (a *astBooleanNot) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
	val, err := a.Operand.evalItem(ctx, item)
	if err != nil {
		return nil, err
	}

	if !a.HasNot {
		return val, nil
	}

	return &types.AttributeValueMemberBOOL{Value: !isAttributeTrue(val)}, nil
}

func (a *astBooleanNot) canModifyItem(ctx *evalContext, item models.Item) bool {
	if a.HasNot {
		return false
	}
	return a.Operand.canModifyItem(ctx, item)
}

func (a *astBooleanNot) setEvalItem(ctx *evalContext, item models.Item, value types.AttributeValue) error {
	if a.HasNot {
		return PathNotSettableError{}
	}
	return a.Operand.setEvalItem(ctx, item, value)
}

func (a *astBooleanNot) deleteAttribute(ctx *evalContext, item models.Item) error {
	if a.HasNot {
		return PathNotSettableError{}
	}
	return a.Operand.deleteAttribute(ctx, item)
}

func (d *astBooleanNot) String() string {
	sb := new(strings.Builder)
	if d.HasNot {
		sb.WriteString(" not ")
	}
	sb.WriteString(d.Operand.String())
	return sb.String()
}

type irBoolNot struct {
	atom irAtom
}

func (d *irBoolNot) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	cb, err := d.atom.calcQueryForScan(info)
	if err != nil {
		return expression.ConditionBuilder{}, err
	}

	return expression.Not(cb), nil
}
