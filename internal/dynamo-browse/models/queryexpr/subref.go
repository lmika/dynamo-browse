package queryexpr

import (
	"github.com/lmika/dynamo-browse/internal/common/sliceutils"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
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

	quals := make([]any, 0)
	for _, sr := range r.SubRefs {
		sv, err := sr.evalToStrOrInt(ctx, nil)
		if err != nil {
			return nil, err
		}
		quals = append(quals, sv)
	}
	return irNamePath{name: namePath.name, quals: quals}, nil
}

func (r *astSubRef) evalItem(ctx *evalContext, item models.Item) (exprValue, error) {
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

func (r *astSubRef) evalSubRefs(ctx *evalContext, item models.Item, res exprValue, subRefs []*astSubRefType) (exprValue, error) {
	for i, sr := range subRefs {
		sv, err := sr.evalToStrOrInt(ctx, nil)
		if err != nil {
			return nil, err
		}

		switch val := sv.(type) {
		case string:
			mapRes, isMapRes := res.(mappableExprValue)
			if !isMapRes {
				return nil, newValueNotAMapError(r, subRefs[:i+1])
			}

			if mapRes.hasKey(val) {
				res, err = mapRes.valueOf(val)
				if err != nil {
					return nil, err
				}
			} else {
				res = nil
			}
		case int64:
			listRes, isMapRes := res.(slicableExprValue)
			if !isMapRes {
				return nil, newValueNotAListError(r, subRefs[:i+1])
			}

			// TODO - deal with index properly (i.e. error handling)
			res, err = listRes.valueAt(int(val))
			if err != nil {
				return nil, err
			}
		}
	}
	return res, nil
}

func (r *astSubRef) canModifyItem(ctx *evalContext, item models.Item) bool {
	return r.Ref.canModifyItem(ctx, item)
}

func (r *astSubRef) setEvalItem(ctx *evalContext, item models.Item, value exprValue) error {
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
		mapRes, isMapRes := parentItem.(modifiableMapExprValue)
		if !isMapRes {
			return newValueNotAMapError(r, r.SubRefs)
		}

		mapRes.setValueOf(val, value)
	case int64:
		listRes, isMapRes := parentItem.(modifiableSliceExprValue)
		if !isMapRes {
			return newValueNotAListError(r, r.SubRefs)
		}

		listRes.setValueAt(int(val), value)
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
		mapRes, isMapRes := parentItem.(modifiableMapExprValue)
		if !isMapRes {
			return newValueNotAMapError(r, r.SubRefs)
		}

		mapRes.deleteValueOf(val)
	case int64:
		listRes, isMapRes := parentItem.(modifiableSliceExprValue)
		if !isMapRes {
			return newValueNotAListError(r, r.SubRefs)
		}

		// TODO: handle indexes out of bounds
		listRes.deleteValueAt(int(val))
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
	case stringableExprValue:
		return v.asString(), nil
	case numberableExprValue:
		return v.asInt(), nil
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
