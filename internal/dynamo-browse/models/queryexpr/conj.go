package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"strings"
)

func (a *astConjunction) evalToIR(tableInfo *models.TableInfo) (*irConjunction, error) {
	atoms := make([]*irBoolNot, len(a.Operands))
	for i, op := range a.Operands {
		var err error
		atoms[i], err = op.evalToIR(tableInfo)
		if err != nil {
			return nil, err
		}
	}

	return &irConjunction{atoms: atoms}, nil
}

func (a *astConjunction) evalItem(item models.Item) (types.AttributeValue, error) {
	val, err := a.Operands[0].evalItem(item)
	if err != nil {
		return nil, err
	}
	if len(a.Operands) == 1 {
		return val, nil
	}

	for _, opr := range a.Operands[1:] {
		if !isAttributeTrue(val) {
			return &types.AttributeValueMemberBOOL{Value: false}, nil
		}

		val, err = opr.evalItem(item)
		if err != nil {
			return nil, err
		}
	}

	return &types.AttributeValueMemberBOOL{Value: isAttributeTrue(val)}, nil
}

func (d *astConjunction) String() string {
	sb := new(strings.Builder)
	for i, operand := range d.Operands {
		if i > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(operand.String())
	}
	return sb.String()
}

type irConjunction struct {
	atoms []*irBoolNot
}

func (d *irConjunction) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	switch len(d.atoms) {
	case 1:
		return d.atoms[0].operandFieldName() == info.Keys.PartitionKey && d.atoms[0].canBeExecutedAsQuery(info, qci)
	case 2:
		return d.atoms[0].canBeExecutedAsQuery(info, qci) && d.atoms[1].canBeExecutedAsQuery(info, qci)
	}
	return false
}

func (d *irConjunction) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	if len(d.atoms) == 1 {
		return d.atoms[0].calcQueryForQuery(info)
	} else if len(d.atoms) != 2 {
		return expression.KeyConditionBuilder{}, errors.Errorf("internal error: expected len to be either 1 or 2, but was %v", len(d.atoms))
	}

	left, err := d.atoms[0].calcQueryForQuery(info)
	if err != nil {
		return expression.KeyConditionBuilder{}, err
	}

	right, err := d.atoms[1].calcQueryForQuery(info)
	if err != nil {
		return expression.KeyConditionBuilder{}, err
	}

	if d.atoms[0].operandFieldName() == info.Keys.PartitionKey {
		return expression.KeyAnd(left, right), nil
	}
	return expression.KeyAnd(right, left), nil
}

func (d *irConjunction) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	if len(d.atoms) == 1 {
		return d.atoms[0].calcQueryForScan(info)
	}

	// TODO: check if can be query
	conds := make([]expression.ConditionBuilder, len(d.atoms))
	for i, operand := range d.atoms {
		cond, err := operand.calcQueryForScan(info)
		if err != nil {
			return expression.ConditionBuilder{}, err
		}
		conds[i] = cond
	}

	// Build conjunction
	conjExpr := expression.And(conds[0], conds[1], conds[2:]...)
	return conjExpr, nil
}

func isAttributeTrue(attr types.AttributeValue) bool {
	switch val := attr.(type) {
	case *types.AttributeValueMemberS:
		return val.Value != ""
	case *types.AttributeValueMemberN:
		return val.Value != "0"
	case *types.AttributeValueMemberBOOL:
		return val.Value
	case *types.AttributeValueMemberNULL:
		return false
	}
	return true
}
