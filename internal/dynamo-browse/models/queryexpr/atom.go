package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

func (a *astAtom) evalToIR(info *models.TableInfo) (irAtom, error) {
	switch {
	case a.Ref != nil:
		return a.Ref.evalToIR(info)
	case a.Literal != nil:
		return a.Literal.evalToIR(info)
	case a.Paren != nil:
		return a.Paren.evalToIR(info)
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
	switch {
	case a.Ref != nil:
		return a.Ref.evalItem(item)
	case a.Literal != nil:
		return a.Literal.dynamoValue()
	case a.Paren != nil:
		return a.Paren.evalItem(item)
	}

	return nil, errors.New("unhandled atom case")
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
