package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

const (
	valuePlaceholderPrefix = '$'
	namePlaceholderPrefix  = ':'
)

func (p *astPlaceholder) evalToIR(ctx *evalContext, info *models.TableInfo) (irAtom, error) {
	placeholderType := p.Placeholder[0]
	placeholder := p.Placeholder[1:]

	if placeholderType == valuePlaceholderPrefix {
		val, hasVal := ctx.lookupValue(placeholder)
		if !hasVal {
			return nil, MissingPlaceholderError{Placeholder: p.Placeholder}
		}

		return irValue{value: val}, nil
	} else if placeholderType == namePlaceholderPrefix {
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

	if placeholderType == valuePlaceholderPrefix {
		val, hasVal := ctx.lookupValue(placeholder)
		if !hasVal {
			return nil, MissingPlaceholderError{Placeholder: p.Placeholder}
		}
		return val, nil
	} else if placeholderType == namePlaceholderPrefix {
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

func (p *astPlaceholder) canModifyItem(ctx *evalContext, item models.Item) bool {
	placeholderType := p.Placeholder[0]
	return placeholderType == namePlaceholderPrefix
}

func (p *astPlaceholder) setEvalItem(ctx *evalContext, item models.Item, value types.AttributeValue) error {
	placeholderType := p.Placeholder[0]
	placeholder := p.Placeholder[1:]

	if placeholderType == valuePlaceholderPrefix {
		return PathNotSettableError{}
	} else if placeholderType == namePlaceholderPrefix {
		name, hasName := ctx.lookupName(placeholder)
		if !hasName {
			return MissingPlaceholderError{Placeholder: p.Placeholder}
		}

		item[name] = value
		return nil
	}

	return errors.New("unrecognised placeholder")
}

func (p *astPlaceholder) deleteAttribute(ctx *evalContext, item models.Item) error {
	placeholderType := p.Placeholder[0]
	placeholder := p.Placeholder[1:]

	if placeholderType == valuePlaceholderPrefix {
		return PathNotSettableError{}
	} else if placeholderType == namePlaceholderPrefix {
		name, hasName := ctx.lookupName(placeholder)
		if !hasName {
			return MissingPlaceholderError{Placeholder: p.Placeholder}
		}

		delete(item, name)
		return nil
	}

	return errors.New("unrecognised placeholder")
}

func (p *astPlaceholder) String() string {
	return p.Placeholder
}
