package queryexpr

import (
	"github.com/alecthomas/participle/v2"
	"github.com/pkg/errors"
)

// Modelled on the expression language here
// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.OperatorsAndFunctions.html

type astExpr struct {
	Root *astDisjunction `parser:"@@"`
}

type astDisjunction struct {
	Operands []*astBinOp `parser:"@@ ('or' @@)*"`
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
