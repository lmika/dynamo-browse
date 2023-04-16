package queryexpr

import (
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

func (a *astAtom) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	switch {
	case a.Ref != nil:
		return a.Ref.evalToIR(ctx, info)
	case a.Literal != nil:
		return a.Literal.evalToIR(ctx, info)
	case a.Placeholder != nil:
		return a.Placeholder.evalToIR(ctx, info)
	case a.Paren != nil:
		return a.Paren.evalToIR(ctx, info)
	}

	return nil, errors.New("unhandled atom case")
}

func (a *astAtom) unqualifiedName() (string, bool) {
	switch {
	case a.Ref != nil:
		return a.Ref.unqualifiedName()
	}

	return "", false
}

func (a *astAtom) evalItem(ctx *evalContext, item models.Item) (exprValue, error) {
	switch {
	case a.Ref != nil:
		return a.Ref.evalItem(ctx, item)
	case a.Literal != nil:
		return a.Literal.exprValue()
	case a.Placeholder != nil:
		return a.Placeholder.evalItem(ctx, item)
	case a.Paren != nil:
		return a.Paren.evalItem(ctx, item)
	}

	return nil, errors.New("unhandled atom case")
}

func (a *astAtom) canModifyItem(ctx *evalContext, item models.Item) bool {
	switch {
	case a.Ref != nil:
		return a.Ref.canModifyItem(ctx, item)
	case a.Placeholder != nil:
		return a.Placeholder.canModifyItem(ctx, item)
	case a.Paren != nil:
		return a.Paren.canModifyItem(ctx, item)
	}
	return false
}

func (a *astAtom) setEvalItem(ctx *evalContext, item models.Item, value exprValue) error {
	switch {
	case a.Ref != nil:
		return a.Ref.setEvalItem(ctx, item, value)
	case a.Placeholder != nil:
		return a.Placeholder.setEvalItem(ctx, item, value)
	case a.Paren != nil:
		return a.Paren.setEvalItem(ctx, item, value)
	}
	return PathNotSettableError{}
}

func (a *astAtom) deleteAttribute(ctx *evalContext, item models.Item) error {
	switch {
	case a.Ref != nil:
		return a.Ref.deleteAttribute(ctx, item)
	case a.Paren != nil:
		return a.Paren.deleteAttribute(ctx, item)
	case a.Placeholder != nil:
		return a.Placeholder.deleteAttribute(ctx, item)
	}
	return PathNotSettableError{}
}

func (a *astAtom) String() string {
	switch {
	case a.Ref != nil:
		return a.Ref.String()
	case a.Literal != nil:
		return a.Literal.String()
	case a.Paren != nil:
		return "(" + a.Paren.String() + ")"
	case a.Placeholder != nil:
		return a.Placeholder.String()
	}
	return ""
}
