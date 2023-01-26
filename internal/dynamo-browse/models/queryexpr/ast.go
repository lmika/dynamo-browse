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
	Ref           *astComparisonOp `parser:"@@ ("`
	HasNot        bool             `parser:"@'not'? 'in' "`
	Operand       []*astExpr       `parser:"('(' @@ (',' @@ )*  ')' |"`
	SingleOperand *astComparisonOp `parser:"@@ ))?"`
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
	Ref    *astSubRef `parser:"@@ ( 'is' "`
	HasNot bool       `parser:"@'not'?"`
	Value  *astSubRef `parser:"@@ )?"`
}

type astSubRef struct {
	Ref   *astFunctionCall `parser:"@@"`
	Quals []string         `parser:"('.' @Ident)*"`
}

type astFunctionCall struct {
	Caller *astAtom   `parser:"@@"`
	IsCall bool       `parser:"( @'(' "`
	Args   []*astExpr `parser:"( @@ (',' @@ )*)? ')' )?"`
}

type astAtom struct {
	Ref         *astRef          `parser:"@@ | "`
	Literal     *astLiteralValue `parser:"@@ | "`
	Placeholder *astPlaceholder  `parser:"@@ | "`
	Paren       *astExpr         `parser:"'(' @@ ')'"`
}

type astRef struct {
	Name string `parser:"@Ident"`
}

type astPlaceholder struct {
	Placeholder string `parser:"@PlaceholderIdent"`
}

type astLiteralValue struct {
	StringVal      *string `parser:"@String"`
	IntVal         *int64  `parser:"| @Int"`
	TrueBoolValue  bool    `parser:"| @KwdTrue"`
	FalseBoolValue bool    `parser:"| @KwdFalse"`
}

var scanner = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "KwdTrue", Pattern: `true`},
	{Name: "KwdFalse", Pattern: `false`},
	{Name: "Eq", Pattern: `=|[\\^]=|[!]=`},
	{Name: "Cmp", Pattern: `<[=]?|>[=]?`},
	{Name: "String", Pattern: `"(\\"|[^"])*"`},
	{Name: "Int", Pattern: `[-+]?(\d*\.)?\d+`},
	{Name: "Number", Pattern: `[-+]?(\d*\.)?\d+`},
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_-]*`},
	{Name: "PlaceholderIdent", Pattern: `[$:][a-zA-Z0-9_-][a-zA-Z0-9_-]*`},
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

func (a *astExpr) calcQuery(ctx *evalContext, info *models.TableInfo) (*models.QueryExecutionPlan, error) {
	ir, err := a.evalToIR(ctx, info)
	if err != nil {
		return nil, err
	}

	var qci queryCalcInfo
	if canExecuteAsQuery(ir, info, &qci) {
		ke, err := ir.(queryableIRAtom).calcQueryForQuery(info)
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

func (a *astExpr) evalToIR(ctx *evalContext, tableInfo *models.TableInfo) (irAtom, error) {
	return a.Root.evalToIR(ctx, tableInfo)
}

func (a *astExpr) evalItem(ctx *evalContext, item models.Item) (types.AttributeValue, error) {
	return a.Root.evalItem(ctx, item)
}

func (a *astExpr) setEvalItem(ctx *evalContext, item models.Item, value types.AttributeValue) error {
	return a.Root.setEvalItem(ctx, item, value)
}

func (a *astExpr) deleteAttribute(ctx *evalContext, item models.Item) error {
	return a.Root.deleteAttribute(ctx, item)
}

func (md *astExpr) canModifyItem(ctx *evalContext, item models.Item) bool {
	return md.Root.canModifyItem(ctx, item)
}
