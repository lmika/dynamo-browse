package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"strings"
)

func (dt *astDot) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	return irNamePath{dt.Name, dt.Quals}, nil
}

func (dt *astDot) unqualifiedName() (string, bool) {
	if len(dt.Quals) == 0 {
		return dt.Name, true
	}
	return "", false
}

func (dt *astDot) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
	res, hasV := item[dt.Name]
	if !hasV {
		return nil, nil
	}

	for i, qualName := range dt.Quals {
		mapRes, isMapRes := res.(*types.AttributeValueMemberM)
		if !isMapRes {
			return nil, ValueNotAMapError(append([]string{dt.Name}, dt.Quals[:i+1]...))
		}

		res, hasV = mapRes.Value[qualName]
		if !hasV {
			return nil, nil
		}
	}

	return res, nil
}

func (dt *astDot) canModifyItem(item models.Item) bool {
	return true
}

func (dt *astDot) setEvalItem(item models.Item, value types.AttributeValue) error {
	if len(dt.Quals) == 0 {
		item[dt.Name] = value
		return nil
	}

	parentItem := item[dt.Name]
	for i, key := range dt.Quals {
		mapItem, isMapItem := parentItem.(*types.AttributeValueMemberM)
		if !isMapItem {
			return PathNotSettableError{}
		}

		if isLast := i == len(dt.Quals)-1; isLast {
			mapItem.Value[key] = value
		} else {
			parentItem = mapItem.Value[key]
		}
	}
	return nil
}

func (dt *astDot) deleteAttribute(item models.Item) error {
	if len(dt.Quals) == 0 {
		delete(item, dt.Name)
		return nil
	}

	parentItem := item[dt.Name]
	for i, key := range dt.Quals {
		mapItem, isMapItem := parentItem.(*types.AttributeValueMemberM)
		if !isMapItem {
			return PathNotSettableError{}
		}

		if isLast := i == len(dt.Quals)-1; isLast {
			delete(mapItem.Value, key)
		} else {
			parentItem = mapItem.Value[key]
		}
	}
	return nil
}

func (a *astDot) String() string {
	var sb strings.Builder

	sb.WriteString(a.Name)
	for _, q := range a.Quals {
		sb.WriteRune('.')
		sb.WriteString(q)
	}

	return sb.String()
}

type irNamePath struct {
	name  string
	quals []string
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
	nb := expression.Name(i.name)
	for _, qual := range i.quals {
		nb = nb.AppendName(expression.Name(qual))
	}
	return nb
}
