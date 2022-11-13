package queryexpr

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

// Modelled on the expression language here
// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.OperatorsAndFunctions.html

type astExpr struct {
	Root *astDisjunction `parser:"@@"`
}

type astDisjunction struct {
	Operands []*astConjunction `parser:"@@ ('or' @@)*"`
}

type astConjunction struct {
	Operands []*astBooleanNot `parser:"@@ ('and' @@)*"`
}

type astBooleanNot struct {
	HasNot  bool   `parser:"@'not'? "`
	Operand *astIn `parser:"@@"`
}

type astIn struct {
	Ref     *astComparisonOp `parser:"@@ ("`
	HasNot  bool             `parser:"@'not'? 'in' '('"`
	Operand []*astAtom       `parser:"@@ (',' @@ )*  ')')?"`
}

type astComparisonOp struct {
	Ref   *astEqualityOp `parser:"@@"`
	Op    string         `parser:"( @('<' | '<=' | '>' | '>=')"`
	Value *astEqualityOp `parser:"@@ )?"`
}

type astEqualityOp struct {
	Ref   *astIsOp `parser:"@@"`
	Op    string   `parser:"( @('^=' | '=' | '!=')"`
	Value *astIsOp `parser:"@@ )?"`
}

type astIsOp struct {
	Ref    *astFunctionCall `parser:"@@ ( 'is' "`
	HasNot bool             `parser:"@'not'?"`
	Value  *astFunctionCall `parser:"@@ )?"`
}

type astFunctionCall struct {
	Caller *astAtom   `parser:"@@"`
	IsCall bool       `parser:"( @'(' "`
	Args   []*astExpr `parser:"( @@ (',' @@ )*)? ')' )?"`
}

type astAtom struct {
	Ref     *astDot          `parser:"@@ | "`
	Literal *astLiteralValue `parser:"@@ | "`
	//FnCall  *astFunctionCall `parser:"@@ | "`
	Paren *astExpr `parser:"'(' @@ ')'"`
}

//type astFunctionCall struct {
//	Name string     `parser:"@Ident '('"`
//	Args []*astExpr `parser:"( @@ (',' @@ )*)? ')'"`
//}

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
var parser = participle.MustBuild[astExpr](
	participle.Lexer(scanner),
)

func Parse(expr string) (*QueryExpr, error) {
	ast, err := parser.ParseString("expr", expr)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse expression: '%v'", expr)
	}

	return &QueryExpr{ast: ast}, nil
}

func (a *astExpr) calcQuery(info *models.TableInfo) (*models.QueryExecutionPlan, error) {
	ir, err := a.evalToIR(info)
	if err != nil {
		return nil, err
	}

	var qci queryCalcInfo
	if ir.canBeExecutedAsQuery(info, &qci) {
		ke, err := ir.calcQueryForQuery(info)
		if err != nil {
			return nil, err
		}

		builder := expression.NewBuilder()
		builder = builder.WithKeyCondition(ke)

		expr, err := builder.Build()
		if err != nil {
			return nil, err
		}

		return &models.QueryExecutionPlan{
			CanQuery:   true,
			Expression: expr,
		}, nil
	}

	cb, err := ir.calcQueryForScan(info)
	if err != nil {
		return nil, err
	}

	builder := expression.NewBuilder()
	builder = builder.WithFilter(cb)

	expr, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return &models.QueryExecutionPlan{
		CanQuery:   false,
		Expression: expr,
	}, nil
}

func (a *astExpr) evalToIR(tableInfo *models.TableInfo) (irAtom, error) {
	return a.Root.evalToIR(tableInfo)
}

func (a *astExpr) evalItem(item models.Item) (types.AttributeValue, error) {
	return a.Root.evalItem(item)
}
