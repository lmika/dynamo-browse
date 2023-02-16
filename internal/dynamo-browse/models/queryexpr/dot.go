package queryexpr

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"strings"
)

func (dt *astRef) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	return irNamePath{name: dt.Name}, nil
}

func (dt *astRef) unqualifiedName() (string, bool) {
	return dt.Name, true
}

func (dt *astRef) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
	res, hasV := item[dt.Name]
	if !hasV {
		return nil, nil
	}

	return res, nil
}

func (dt *astRef) canModifyItem(ctx *evalContext, item models.Item) bool {
	return true
}

func (dt *astRef) setEvalItem(ctx *evalContext, item models.Item, value types.AttributeValue) error {
	item[dt.Name] = value
	return nil
}

func (dt *astRef) deleteAttribute(ctx *evalContext, item models.Item) error {
	delete(item, dt.Name)
	return nil
}

func (a *astRef) String() string {
	return a.Name
}

type irNamePath struct {
	name  string
	quals []any
}

func (i irNamePath) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	return expression.ConditionBuilder{}, NodeCannotBeConvertedToQueryError{}
}

func (i irNamePath) calcOperand(info *models.TableInfo) expression.OperandBuilder {
	return i.calcName(info)
}

func (i irNamePath) keyName() string {
	if len(i.quals) > 0 {
		return ""
	}
	return i.name
}

func (i irNamePath) calcName(info *models.TableInfo) expression.NameBuilder {
	var fullName strings.Builder
	fullName.WriteString(i.name)

	for _, qual := range i.quals {
		switch v := qual.(type) {
		case string:
			fullName.WriteString("." + v)
		case int:
			fullName.WriteString(fmt.Sprintf("[%v]", qual))
		}
	}
	return expression.NameNoDotSplit(fullName.String())
}
