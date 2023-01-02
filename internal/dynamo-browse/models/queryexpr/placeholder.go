package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

func (p *astPlaceholder) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	placeholderType := p.Placeholder[0]
	placeholder := p.Placeholder[1:]

	if placeholderType == '$' {
		val, hasVal := ctx.lookupValue(placeholder)
		if !hasVal {
			return nil, MissingPlaceholderError{Placeholder: p.Placeholder}
		}

		return irValue{value: val}, nil
	} else if placeholderType == ':' {
		name, hasName := ctx.lookupName(placeholder)
		if !hasName {
			return nil, MissingPlaceholderError{Placeholder: p.Placeholder}
		}

		return irNamePath{name, nil}, nil
	}

	return nil, errors.New("unrecognised placeholder")
}

func (p *astPlaceholder) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
	placeholderType := p.Placeholder[0]
	placeholder := p.Placeholder[1:]

	if placeholderType == '$' {
		val, hasVal := ctx.lookupValue(placeholder)
		if !hasVal {
			return nil, MissingPlaceholderError{Placeholder: p.Placeholder}
		}
		return val, nil
	} else if placeholderType == ':' {
		name, hasName := ctx.lookupName(placeholder)
		if !hasName {
			return nil, MissingPlaceholderError{Placeholder: p.Placeholder}
		}

		res, hasV := item[name]
		if !hasV {
			return nil, nil
		}

		return res, nil
	}

	return nil, errors.New("unrecognised placeholder")
}
