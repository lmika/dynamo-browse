package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

func (a *astAtom) evalToIR(info *models.TableInfo) (irAtom, error) {
	switch {
	case a.Ref != nil:
		return nil, errors.New("TODO")
	case a.Literal != nil:
		return nil, errors.New("TODO")
	case a.Paren != nil:
		return a.Paren.evalToIR(info)
	}

	return nil, errors.New("unhandled atom case")
}

func (a *astAtom) rightOperandGoValue() (any, error) {
	switch {
	case a.Ref != nil:
		return nil, errors.New("literal value required")
	case a.Literal != nil:
		return a.Literal.goValue()
	case a.Paren != nil:
		return nil, errors.New("TODO")
	}

	return nil, errors.New("unhandled atom case")
}

func (a *astAtom) rightOperandDynamoValue() (types.AttributeValue, error) {
	switch {
	case a.Literal != nil:
		return a.Literal.dynamoValue()
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

func (a *astAtom) evalItem(item models.Item) (types.AttributeValue, error) {
	return nil, errors.New("TODO")
}

func (a *astAtom) String() string {
	switch {
	case a.Ref != nil:
		return a.Ref.String()
	case a.Literal != nil:
		return a.Literal.String()
	case a.Paren != nil:
		return "(" + a.Paren.String() + ")"
	}
	return ""
}
