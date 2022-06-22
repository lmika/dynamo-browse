package queryexpr

import (
	"github.com/alecthomas/participle/v2"
	"github.com/pkg/errors"
)

type astExpr struct {
	Equality *astBinOp `parser:"@@"`
}

type astBinOp struct {
	Name  string           `parser:"@Ident"`
	Op    string           `parser:"@('^' '=' | '=')"`
	Value *astLiteralValue `parser:"@@"`
}

type astLiteralValue struct {
	StringVal string `parser:"@String"`
}

var parser = participle.MustBuild(&astExpr{})

func Parse(expr string) (*QueryExpr, error) {
	var ast astExpr

	if err := parser.ParseString("expr", expr, &ast); err != nil {
		return nil, errors.Wrapf(err, "cannot parse expression: '%v'", expr)
	}

	return &QueryExpr{ast: &ast}, nil
}
