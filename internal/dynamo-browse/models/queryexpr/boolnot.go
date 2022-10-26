package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"strings"
)

func (a *astBooleanNot) evalToIR(tableInfo *models.TableInfo) (*irBoolNot, error) {
	irNode, err := a.Operand.evalToIR(tableInfo)
	if err != nil {
		return nil, err
	}

	return &irBoolNot{hasNot: a.HasNot, atom: irNode}, nil
}

func (a *astBooleanNot) evalItem(item models.Item) (types.AttributeValue, error) {
	return nil, errors.New("todo")
	//val, err := a.Operands[0].evalItem(item)
	//if err != nil {
	//	return nil, err
	//}
	//if len(a.Operands) == 1 {
	//	return val, nil
	//}
	//
	//for _, opr := range a.Operands[1:] {
	//	if !isAttributeTrue(val) {
	//		return &types.AttributeValueMemberBOOL{Value: false}, nil
	//	}
	//
	//	val, err = opr.evalItem(item)
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	//
	//return &types.AttributeValueMemberBOOL{Value: isAttributeTrue(val)}, nil
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
	hasNot bool
	atom   irAtom
}

func (d *irBoolNot) operandFieldName() string {
	if d.hasNot {
		return ""
	}
	return d.atom.operandFieldName()
}

func (d *irBoolNot) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	if d.hasNot {
		return false
	}
	return d.atom.canBeExecutedAsQuery(info, qci)
}

func (d *irBoolNot) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	if d.hasNot {
		return expression.KeyConditionBuilder{}, errors.New("query not supported")
	}

	return d.atom.calcQueryForQuery(info)
}

func (d *irBoolNot) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	cb, err := d.atom.calcQueryForScan(info)
	if err != nil {
		return expression.ConditionBuilder{}, err
	}

	if d.hasNot {
		return expression.Not(cb), nil
	}
	return cb, nil
}
