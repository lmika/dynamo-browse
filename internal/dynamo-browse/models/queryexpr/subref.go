package queryexpr

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/common/sliceutils"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"strconv"
	"strings"
)

func (r *astSubRef) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	refIR, err := r.Ref.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}
	if len(r.SubRefs) == 0 {
		return refIR, nil
	}

	// This node has subrefs
	namePath, isNamePath := refIR.(irNamePath)
	if !isNamePath {
		return nil, OperandNotANameError(r.String())
	}

	quals := make([]string, 0)
	for _, sr := range r.SubRefs {
		sv, err := sr.evalToStrOrInt(ctx, nil)
		if err != nil {
			return nil, err
		}

		switch val := sv.(type) {
		case string:
			quals = append(quals, val)
		case int:
			quals = append(quals, fmt.Sprintf("[%v]", val))
		}
	}
	return irNamePath{name: namePath.name, quals: quals}, nil
}

func (r *astSubRef) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
	res, err := r.Ref.evalItem(ctx, item)
	if err != nil {
		return nil, err
	}

	res, err = r.evalSubRefs(ctx, item, res, r.SubRefs)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *astSubRef) evalSubRefs(ctx *evalContext, item models.Item, res types.AttributeValue, subRefs []*astSubRefType) (types.AttributeValue, error) {
	for i, sr := range subRefs {
		sv, err := sr.evalToStrOrInt(ctx, nil)
		if err != nil {
			return nil, err
		}

		switch val := sv.(type) {
		case string:
			var hasV bool
			mapRes, isMapRes := res.(*types.AttributeValueMemberM)
			if !isMapRes {
				return nil, newValueNotAMapError(r, subRefs[:i+1])
			}

			res, hasV = mapRes.Value[val]
			if !hasV {
				return nil, nil
			}
		case int:
			listRes, isMapRes := res.(*types.AttributeValueMemberL)
			if !isMapRes {
				return nil, newValueNotAListError(r, subRefs[:i+1])
			}

			// TODO - deal with index properly
			res = listRes.Value[val]
		}
	}
	return res, nil
}

func (r *astSubRef) canModifyItem(ctx *evalContext, item models.Item) bool {
	return r.Ref.canModifyItem(ctx, item)
}

func (r *astSubRef) setEvalItem(ctx *evalContext, item models.Item, value types.AttributeValue) error {
	if len(r.SubRefs) == 0 {
		return r.Ref.setEvalItem(ctx, item, value)
	}

	parentItem, err := r.Ref.evalItem(ctx, item)
	if err != nil {
		return err
	}

	if len(r.SubRefs) > 1 {
		parentItem, err = r.evalSubRefs(ctx, item, parentItem, r.SubRefs[0:len(r.SubRefs)-1])
		if err != nil {
			return err
		}
	}

	sv, err := r.SubRefs[len(r.SubRefs)-1].evalToStrOrInt(ctx, nil)
	if err != nil {
		return err
	}

	switch val := sv.(type) {
	case string:
		mapRes, isMapRes := parentItem.(*types.AttributeValueMemberM)
		if !isMapRes {
			return newValueNotAMapError(r, r.SubRefs)
		}

		mapRes.Value[val] = value
	case int:
		listRes, isMapRes := parentItem.(*types.AttributeValueMemberL)
		if !isMapRes {
			return newValueNotAListError(r, r.SubRefs)
		}

		// TODO: handle indexes
		listRes.Value[val] = value
	}
	return nil
}

func (r *astSubRef) deleteAttribute(ctx *evalContext, item models.Item) error {
	if len(r.SubRefs) == 0 {
		return r.Ref.deleteAttribute(ctx, item)
	}

	parentItem, err := r.Ref.evalItem(ctx, item)
	if err != nil {
		return err
	}

	/*
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
	*/
	if len(r.SubRefs) > 1 {
		parentItem, err = r.evalSubRefs(ctx, item, parentItem, r.SubRefs[0:len(r.SubRefs)-1])
		if err != nil {
			return err
		}
	}

	sv, err := r.SubRefs[len(r.SubRefs)-1].evalToStrOrInt(ctx, nil)
	if err != nil {
		return err
	}

	switch val := sv.(type) {
	case string:
		mapRes, isMapRes := parentItem.(*types.AttributeValueMemberM)
		if !isMapRes {
			return newValueNotAMapError(r, r.SubRefs)
		}

		delete(mapRes.Value, val)
	case int:
		listRes, isMapRes := parentItem.(*types.AttributeValueMemberL)
		if !isMapRes {
			return newValueNotAListError(r, r.SubRefs)
		}

		// TODO: handle indexes out of bounds
		oldList := listRes.Value
		newList := append([]types.AttributeValue{}, oldList[:val]...)
		newList = append(newList, oldList[val+1:]...)
		listRes.Value = newList
	}
	return nil
}

func (r *astSubRef) String() string {
	var sb strings.Builder

	sb.WriteString(r.Ref.String())
	for _, q := range r.SubRefs {
		switch {
		case q.DotQual != "":
			sb.WriteRune('.')
			sb.WriteString(q.DotQual)
		case q.SubIndex != nil:
			sb.WriteRune('[')
			sb.WriteString(q.SubIndex.String())
			sb.WriteRune(']')
		}
	}

	return sb.String()
}

func (sr *astSubRefType) evalToStrOrInt(ctx *evalContext, item models.Item) (any, error) {
	if sr.DotQual != "" {
		return sr.DotQual, nil
	}

	subEvalItem, err := sr.SubIndex.evalItem(ctx, item)
	if err != nil {
		return nil, err
	}
	switch v := subEvalItem.(type) {
	case *types.AttributeValueMemberS:
		return v.Value, nil
	case *types.AttributeValueMemberN:
		intVal, err := strconv.Atoi(v.Value)
		if err == nil {
			return intVal, nil
		}
		flVal, err := strconv.ParseFloat(v.Value, 64)
		if err == nil {
			return int(flVal), nil
		}
		return nil, err
	}
	return nil, ValueNotUsableAsASubref{}
}

func (sr *astSubRefType) string() string {
	switch {
	case sr.DotQual != "":
		return sr.DotQual
	case sr.SubIndex != nil:
		return sr.SubIndex.String()
	}
	return ""
}

func newValueNotAMapError(r *astSubRef, subRefs []*astSubRefType) ValueNotAMapError {
	subRefStrings := sliceutils.Map(subRefs, func(srt *astSubRefType) string {
		return srt.string()
	})
	return ValueNotAMapError(append([]string{r.Ref.String()}, subRefStrings...))
}

func newValueNotAListError(r *astSubRef, subRefs []*astSubRefType) ValueNotAListError {
	subRefStrings := sliceutils.Map(subRefs, func(srt *astSubRefType) string {
		return srt.string()
	})
	return ValueNotAListError(append([]string{r.Ref.String()}, subRefStrings...))
}
