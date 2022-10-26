package queryexpr

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
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
	Operands []*astBooleanNot `parser:"@@ ('and' @@)*"`
}

type astBooleanNot struct {
	HasNot  bool             `parser:"@'not'? "`
	Operand *astComparisonOp `parser:"@@"`
}

type astComparisonOp struct {
	Ref   *astEqualityOp   `parser:"@@"`
	Op    string           `parser:"( @('<' | '<=' | '>' | '>=')"`
	Value *astLiteralValue `parser:"@@ )?"`
}

type astEqualityOp struct {
	Ref   *astDot          `parser:"@@"`
	Op    string           `parser:"( @('^=' | '=' | '!=')"`
	Value *astLiteralValue `parser:"@@ )?"`
}

type astDot struct {
	Name  string   `parser:"@Ident"`
	Quals []string `parser:"('.' @Ident)*"`
}

type astLiteralValue struct {
	StringVal *string `parser:"@String"`
	IntVal    *int64  `parser:"| @Int"`
}

var scanner = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Eq", Pattern: `=|[\\^]=|[!]=`},
	{Name: "Cmp", Pattern: `<[=]?|>[=]?`},
	{Name: "String", Pattern: `"(\\"|[^"])*"`},
	{Name: "Int", Pattern: `[-+]?(\d*\.)?\d+`},
	{Name: "Number", Pattern: `[-+]?(\d*\.)?\d+`},
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_-]*`},
	{Name: "Punct", Pattern: `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|][=]?`},
	{Name: "EOL", Pattern: `[\n\r]+`},
	{Name: "whitespace", Pattern: `[ \t]+`},
})
var parser = participle.MustBuild[astExpr](participle.Lexer(scanner))

func Parse(expr string) (*QueryExpr, error) {
	ast, err := parser.ParseString("expr", expr)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse expression: '%v'", expr)
	}

	return &QueryExpr{ast: ast}, nil
}
