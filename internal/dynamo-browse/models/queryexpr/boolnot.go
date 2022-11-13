package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"strings"
)

func (a *astBooleanNot) evalToIR(tableInfo *models.TableInfo) (irAtom, error) {
	irNode, err := a.Operand.evalToIR(tableInfo)
	if err != nil {
		return nil, err
	}

	if !a.HasNot {
		return irNode, nil
	}

	return &irBoolNot{atom: irNode}, nil
}

func (a *astBooleanNot) evalItem(item models.Item) (types.AttributeValue, error) {
	return nil, errors.New("todo")
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
