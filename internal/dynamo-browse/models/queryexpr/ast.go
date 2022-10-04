package queryexpr

import (
	"github.com/alecthomas/participle/v2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

// Modelled on the expression language here
// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.OperatorsAndFunctions.html

type astExpr struct {
	Root *astDisjunction `parser:"@@"`
}

func (a *astExpr) evalToIR(tableInfo *models.TableInfo) (*irDisjunction, error) {
	return a.Root.evalToIR(tableInfo)
}

func (a *astExpr) evalItem(item models.Item) (types.AttributeValue, error) {
	return a.Root.evalItem(item)
}

type astDisjunction struct {
	Operands []*astConjunction `parser:"@@ ('or' @@)*"`
}

type astConjunction struct {
	Operands []*astBinOp `parser:"@@ ('and' @@)*"`
}

// TODO: do this properly
type astBinOp struct {
	Ref   *astDot          `parser:"@@"`
	Op    string           `parser:"( @('^' '=' | '=')"`
	Value *astLiteralValue `parser:"@@ )?"`
}

type astDot struct {
	Name  string   `parser:"@Ident"`
	Quals []string `parser:"('.' @Ident)*"`
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
