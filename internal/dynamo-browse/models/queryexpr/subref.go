package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"strings"
)

func (r *astSubRef) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	refIR, err := r.Ref.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}
	if len(r.Quals) == 0 {
		return refIR, nil
	}

	// This node has subrefs
	namePath, isNamePath := refIR.(irNamePath)
	if !isNamePath {
		return nil, OperandNotANameError(r.String())
	}

	quals := make([]string, 0)
	for _, sr := range r.Quals {
		quals = append(quals, sr)
	}
	return irNamePath{name: namePath.name, quals: quals}, nil
}

func (r *astSubRef) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
	res, err := r.Ref.evalItem(ctx, item)
	if err != nil {
		return nil, err
	}

	for i, qualName := range r.Quals {
		var hasV bool

		mapRes, isMapRes := res.(*types.AttributeValueMemberM)
		if !isMapRes {
			return nil, ValueNotAMapError(append([]string{r.Ref.String()}, r.Quals[:i+1]...))
		}

		res, hasV = mapRes.Value[qualName]
		if !hasV {
			return nil, nil
		}
	}

	return res, nil
}

func (r *astSubRef) canModifyItem(ctx *evalContext, item models.Item) bool {
	return r.Ref.canModifyItem(ctx, item)
}

func (r *astSubRef) setEvalItem(ctx *evalContext, item models.Item, value types.AttributeValue) error {
	if len(r.Quals) == 0 {
		return r.Ref.setEvalItem(ctx, item, value)
	}

	parentItem, err := r.Ref.evalItem(ctx, item)
	if err != nil {
		return err
	}

	for i, key := range r.Quals {
		mapItem, isMapItem := parentItem.(*types.AttributeValueMemberM)
		if !isMapItem {
			return PathNotSettableError{}
		}

		if isLast := i == len(r.Quals)-1; isLast {
			mapItem.Value[key] = value
		} else {
			parentItem = mapItem.Value[key]
		}
	}
	return nil
}

func (r *astSubRef) deleteAttribute(ctx *evalContext, item models.Item) error {
	if len(r.Quals) == 0 {
		return r.Ref.deleteAttribute(ctx, item)
	}

	parentItem, err := r.Ref.evalItem(ctx, item)
	if err != nil {
		return err
	}

	for i, key := range r.Quals {
		mapItem, isMapItem := parentItem.(*types.AttributeValueMemberM)
		if !isMapItem {
			return PathNotSettableError{}
		}

		if isLast := i == len(r.Quals)-1; isLast {
			delete(mapItem.Value, key)
		} else {
			parentItem = mapItem.Value[key]
		}
	}
	return nil
}

func (r *astSubRef) String() string {
	var sb strings.Builder

	sb.WriteString(r.Ref.String())
	for _, q := range r.Quals {
		sb.WriteRune('.')
		sb.WriteString(q)
	}

	return sb.String()
}
