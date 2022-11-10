package queryexpr_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models/queryexpr"
	"testing"

	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/stretchr/testify/assert"
)

func TestModExpr_Query(t *testing.T) {
	tableInfo := &models.TableInfo{
		Name: "test",
		Keys: models.KeyAttribute{
			PartitionKey: "pk",
			SortKey:      "sk",
		},
	}

	t.Run("as queries", func(t *testing.T) {
		scenarios := []scanScenario{
			scanCase("when request pk is fixed",
				`pk="prefix"`,
				`#0 = :0`,
				exprNameIsString(0, 0, "pk", "prefix"),
			),
			scanCase("when request pk is fixed in parens #1",
				`(pk="prefix")`,
				`#0 = :0`,
				exprNameIsString(0, 0, "pk", "prefix"),
			),
			scanCase("when request pk is fixed in parens #2",
				`(pk)="prefix"`,
				`#0 = :0`,
				exprNameIsString(0, 0, "pk", "prefix"),
			),
			scanCase("when request pk is fixed in parens #3",
				`pk=("prefix")`,
				`#0 = :0`,
				exprNameIsString(0, 0, "pk", "prefix"),
			),
			scanCase("when request pk is in with a single value",
				`pk in ("prefix")`,
				`#0 = :0`,
				exprNameIsString(0, 0, "pk", "prefix"),
			),
			scanCase("when request pk and sk is fixed",
				`pk="prefix" and sk="another"`,
				`(#0 = :0) AND (#1 = :1)`,
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsString(1, 1, "sk", "another"),
			),
			scanCase("when request pk and sk is fixed (using 'in')",
				`pk in ("prefix") and sk in ("another")`,
				`(#0 = :0) AND (#1 = :1)`,
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsString(1, 1, "sk", "another"),
			),
			scanCase("when request pk is equals and sk is prefix #1",
				`pk="prefix" and sk^="another"`,
				`(#0 = :0) AND (begins_with (#1, :1))`,
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsString(1, 1, "sk", "another"),
			),
			scanCase("when request pk is equals and sk is prefix #2",
				`sk^="another" and pk="prefix"`,
				`(#0 = :0) AND (begins_with (#1, :1))`,
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsString(1, 1, "sk", "another"),
			),
			scanCase("when request pk is equals and sk is less than",
				`pk="prefix" and sk < 100`,
				`(#0 = :0) AND (#1 < :1)`,
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsNumber(1, 1, "sk", "100"),
			),
			scanCase("when request pk is equals and sk is less or equal to",
				`pk="prefix" and sk <= 100`,
				`(#0 = :0) AND (#1 <= :1)`,
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsNumber(1, 1, "sk", "100"),
			),
			scanCase("when request pk is equals and sk is greater than",
				`pk="prefix" and sk > 100`,
				`(#0 = :0) AND (#1 > :1)`,
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsNumber(1, 1, "sk", "100"),
			),
			scanCase("when request pk is equals and sk is greater or equal to",
				`pk="prefix" and sk >= 100`,
				`(#0 = :0) AND (#1 >= :1)`,
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsNumber(1, 1, "sk", "100"),
			),
		}

		for _, scenario := range scenarios {
			t.Run(scenario.description, func(t *testing.T) {
				modExpr, err := queryexpr.Parse(scenario.expression)
				assert.NoError(t, err)

				plan, err := modExpr.Plan(tableInfo)
				assert.NoError(t, err)

				assert.True(t, plan.CanQuery)
				assert.Equal(t, scenario.expectedFilter, aws.ToString(plan.Expression.KeyCondition()))
				for k, v := range scenario.expectedNames {
					assert.Equal(t, v, plan.Expression.Names()[k])
				}
				for k, v := range scenario.expectedValues {
					assert.Equal(t, v, plan.Expression.Values()[k])
				}
			})
		}
	})

	t.Run("as scans", func(t *testing.T) {
		scenarios := []scanScenario{
			scanCase("when request pk prefix", `pk^="prefix"`, `begins_with (#0, :0)`,
				exprNameIsString(0, 0, "pk", "prefix"),
			),
			scanCase("when request sk equals something", `sk="something"`, `#0 = :0`,
				exprNameIsString(0, 0, "sk", "something"),
			),
			scanCase("with not equal", `sk != "something"`, `#0 <> :0`,
				exprNameIsString(0, 0, "sk", "something"),
			),
			scanCase("less than value", `num < 100`, `#0 < :0`,
				exprNameIsNumber(0, 0, "num", "100"),
			),
			scanCase("less or equal to value", `num <= 100`, `#0 <= :0`,
				exprNameIsNumber(0, 0, "num", "100"),
			),
			scanCase("greater than value", `num > 100`, `#0 > :0`,
				exprNameIsNumber(0, 0, "num", "100"),
			),
			scanCase("greater or equal to value", `num >= 100`, `#0 >= :0`,
				exprNameIsNumber(0, 0, "num", "100"),
			),
			scanCase("with disjunctions",
				`pk="prefix" or sk="another"`,
				`(#0 = :0) OR (#1 = :1)`,
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsString(1, 1, "sk", "another"),
			),
			scanCase("with disjunctions with numbers",
				`pk="prefix" or num=123 and negnum=-131`,
				`(#0 = :0) OR ((#1 = :1) AND (#2 = :2))`,
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsNumber(1, 1, "num", "123"),
				exprNameIsNumber(2, 2, "negnum", "-131"),
			),
			scanCase("with disjunctions with numbers (different priority)",
				`(pk="prefix" or num=123) and negnum=-131`,
				`((#0 = :0) OR (#1 = :1)) AND (#2 = :2)`,
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsNumber(1, 1, "num", "123"),
				exprNameIsNumber(2, 2, "negnum", "-131"),
			),
			scanCase("with disjunctions if pk is present twice in expression",
				`pk="prefix" and pk="another"`,
				`(#0 = :0) AND (#0 = :1)`,
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsString(0, 1, "pk", "another"),
			),
			scanCase("with not", `not pk="prefix"`, `NOT (#0 = :0)`,
				exprNameIsString(0, 0, "pk", "prefix"),
			),

			scanCase("with in", `pk in ("alpha", "bravo", "charlie")`,
				`#0 IN (:0, :1, :2)`,
				exprName(0, "pk"),
				exprValueIsString(0, "alpha"),
				exprValueIsString(1, "bravo"),
				exprValueIsString(2, "charlie"),
			),
			// TODO: in > 100 items ==> items OR items
		}

		for _, scenario := range scenarios {
			t.Run(scenario.description, func(t *testing.T) {
				modExpr, err := queryexpr.Parse(scenario.expression)
				assert.NoError(t, err)

				plan, err := modExpr.Plan(tableInfo)
				assert.NoError(t, err)

				assert.False(t, plan.CanQuery)
				assert.Equal(t, scenario.expectedFilter, aws.ToString(plan.Expression.Filter()))
				for k, v := range scenario.expectedNames {
					assert.Equal(t, v, plan.Expression.Names()[k])
				}
				for k, v := range scenario.expectedValues {
					assert.Equal(t, v, plan.Expression.Values()[k])
				}
			})
		}
	})
}

func TestQueryExpr_EvalItem(t *testing.T) {
	var (
		item = models.Item{
			"alpha": &types.AttributeValueMemberS{Value: "alpha"},
			"bravo": &types.AttributeValueMemberN{Value: "123"},
			"charlie": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"door": &types.AttributeValueMemberS{Value: "red"},
					"tree": &types.AttributeValueMemberS{Value: "green"},
				},
			},
		}
	)

	t.Run("simple values", func(t *testing.T) {
		scenarios := []struct {
			expr     string
			expected types.AttributeValue
		}{
			// Simple values
			{expr: `alpha`, expected: &types.AttributeValueMemberS{Value: "alpha"}},
			{expr: `bravo`, expected: &types.AttributeValueMemberN{Value: "123"}},
			{expr: `charlie`, expected: item["charlie"]},

			// Equality with literal
			{expr: `alpha="alpha"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `bravo=123`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `charlie.tree="green"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha^="al"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha="foobar"`, expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: `alpha^="need-something"`, expected: &types.AttributeValueMemberBOOL{Value: false}},
			// TODO: negation
			// TODO: comparison
			// TODO: in

			// Dot values
			{expr: `charlie.door`, expected: &types.AttributeValueMemberS{Value: "red"}},
			{expr: `charlie.tree`, expected: &types.AttributeValueMemberS{Value: "green"}},

			// Conjunction
			{expr: `alpha="alpha" and bravo=123`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha="alpha" and bravo=321`, expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: `alpha="bravo" and bravo=123`, expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: `alpha="bravo" and bravo=321`, expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: `alpha="alpha" and bravo=123 and charlie.door="red"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha="alpha" and bravo=123 and charlie.door^="green"`, expected: &types.AttributeValueMemberBOOL{Value: false}},

			// Disjunction
			{expr: `alpha="alpha" or bravo=123`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha="alpha" or bravo=321`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha="bravo" or bravo=123`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha="bravo" or bravo=321`, expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: `alpha="alpha" or bravo=123 or charlie.tree="green"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha="bravo" or bravo=321 or charlie.tree^="red"`, expected: &types.AttributeValueMemberBOOL{Value: false}},

			// Order of operation
			{expr: `alpha="alpha" and bravo=123 or charlie.door="green"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha="bravo" or bravo=321 and charlie.door="green"`, expected: &types.AttributeValueMemberBOOL{Value: false}},
		}
		for _, scenario := range scenarios {
			t.Run(scenario.expr, func(t *testing.T) {
				modExpr, err := queryexpr.Parse(scenario.expr)
				assert.NoError(t, err)

				res, err := modExpr.EvalItem(item)
				assert.NoError(t, err)

				assert.Equal(t, scenario.expected, res)
			})
		}
	})

	t.Run("unparsed expression", func(t *testing.T) {
		scenarios := []struct {
			expr          string
			expectedError error
		}{
			{expr: `bla ^ = "something"`},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.expr, func(t *testing.T) {
				_, err := queryexpr.Parse(scenario.expr)
				assert.Error(t, err)
			})
		}
	})

	t.Run("expression errors", func(t *testing.T) {
		scenarios := []struct {
			expr          string
			expectedError error
		}{
			{expr: `not_present`, expectedError: queryexpr.NameNotFoundError("not_present")},
			{expr: `alpha.bravo`, expectedError: queryexpr.ValueNotAMapError([]string{"alpha", "bravo"})},
			{expr: `charlie.tree.bla`, expectedError: queryexpr.ValueNotAMapError([]string{"charlie", "tree", "bla"})},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.expr, func(t *testing.T) {
				modExpr, err := queryexpr.Parse(scenario.expr)
				assert.NoError(t, err)

				res, err := modExpr.EvalItem(item)
				assert.Nil(t, res)
				assert.Equal(t, scenario.expectedError, err)
			})
		}
	})
}

type scanScenario struct {
	description    string
	expression     string
	expectedFilter string
	expectedNames  map[string]string
	expectedValues map[string]types.AttributeValue
}

func scanCase(description, expression, expectedFilter string, options ...func(ss *scanScenario)) scanScenario {
	ss := scanScenario{
		description:    description,
		expression:     expression,
		expectedFilter: expectedFilter,
		expectedNames:  map[string]string{},
		expectedValues: map[string]types.AttributeValue{},
	}
	for _, opt := range options {
		opt(&ss)
	}
	return ss
}

func exprName(idx int, name string) func(ss *scanScenario) {
	return func(ss *scanScenario) {
		ss.expectedNames[fmt.Sprintf("#%d", idx)] = name
	}
}

func exprValueIsString(valIdx int, expected string) func(ss *scanScenario) {
	return func(ss *scanScenario) {
		ss.expectedValues[fmt.Sprintf(":%d", valIdx)] = &types.AttributeValueMemberS{Value: expected}
	}
}

func exprNameIsString(idx, valIdx int, name string, expected string) func(ss *scanScenario) {
	return func(ss *scanScenario) {
		ss.expectedNames[fmt.Sprintf("#%d", idx)] = name
		ss.expectedValues[fmt.Sprintf(":%d", valIdx)] = &types.AttributeValueMemberS{Value: expected}
	}
}

func exprNameIsNumber(idx, valIdx int, name string, expected string) func(ss *scanScenario) {
	return func(ss *scanScenario) {
		ss.expectedNames[fmt.Sprintf("#%d", idx)] = name
		ss.expectedValues[fmt.Sprintf(":%d", valIdx)] = &types.AttributeValueMemberN{Value: expected}
	}
}
