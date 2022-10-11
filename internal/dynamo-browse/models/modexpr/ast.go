package modexpr

import (
	"github.com/alecthomas/participle/v2"
	"github.com/pkg/errors"
)

type astExpr struct {
	Attributes []*astAttribute `parser:"@@ (',' @@)*"`
}

type astAttribute struct {
	Names *astKeyList      `parser:"@@ '='"`
	Value *astLiteralValue `parser:"@@"`
}

type astKeyList struct {
	Names []string `parser:"@Ident ('/' @Ident)*"`
}

type astLiteralValue struct {
	String string `parser:"@String"`
}

var parser = participle.MustBuild[astExpr]()

func Parse(expr string) (*ModExpr, error) {
	ast, err := parser.ParseString("expr", expr)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse expression: '%v'", expr)
	}

	return &ModExpr{ast: ast}, nil
}
