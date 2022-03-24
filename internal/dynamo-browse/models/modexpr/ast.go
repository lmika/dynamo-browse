package modexpr

import (
	"github.com/alecthomas/participle/v2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
	"strconv"
)

type astExpr struct {
	Attributes []*astAttribute `parser:"@@ (',' @@)*"`
}

type astAttribute struct {
	Name  string `parser:"@Ident '='"`
	Value string `parser:"@String"`
}

func (a astAttribute) dynamoValue() (types.AttributeValue, error) {
	// TODO: should be based on type
	s, err := strconv.Unquote(a.Value)
	if err != nil {
		return nil, errors.Wrap(err, "cannot unquote string")
	}
	return &types.AttributeValueMemberS{Value: s}, nil
}


var parser = participle.MustBuild(&astExpr{})

func Parse(expr string) (*ModExpr, error) {
	var ast astExpr

	if err := parser.ParseString("expr", expr, &ast); err != nil {
		return nil, errors.Wrapf(err, "cannot parse expression: '%v'", expr)
	}

	return &ModExpr{ast: &ast}, nil
}