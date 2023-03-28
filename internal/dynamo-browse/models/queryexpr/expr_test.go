package queryexpr_test

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models/queryexpr"
	"testing"
	"time"

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
		GSIs: []models.TableGSI{
			{
				Name: "with-color",
				Keys: models.KeyAttribute{
					PartitionKey: "color",
					SortKey:      "shade",
				},
			},
			{
				Name: "with-apples",
				Keys: models.KeyAttribute{
					PartitionKey: "apples",
					SortKey:      "sk",
				},
			},
			{
				Name: "with-apples-and-oranges",
				Keys: models.KeyAttribute{
					PartitionKey: "apples",
					SortKey:      "oranges",
				},
			},
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

			scanCase("with placeholders",
				`:partition=$valuePrefix and :sort=$valueAnother`,
				`(#0 = :0) AND (#1 = :1)`,
				placeholderNames(map[string]string{
					"partition": "pk",
					"sort":      "sk",
				}),
				placeholderValues(map[string]types.AttributeValue{
					"valuePrefix":  &types.AttributeValueMemberS{Value: "prefix"},
					"valueAnother": &types.AttributeValueMemberS{Value: "another"},
				}),
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsString(1, 1, "sk", "another"),
			),

			// Querying the index
			scanCase("querying the index with the index pk",
				`color="blue"`,
				`#0 = :0`,
				indexName("with-color"),
				exprNameIsString(0, 0, "color", "blue"),
			),
			scanCase("querying the index with the index pk and index sk",
				`color="red" and shade="gray"`,
				`(#0 = :0) AND (#1 = :1)`,
				indexName("with-color"),
				exprNameIsString(0, 0, "color", "red"),
				exprNameIsString(1, 1, "shade", "gray"),
			),
			scanCase("querying the index with the index pk and begins with index sk",
				`color="yellow" and shade ^= "dark"`,
				`(#0 = :0) AND (begins_with (#1, :1))`,
				indexName("with-color"),
				exprNameIsString(0, 0, "color", "yellow"),
				exprNameIsString(1, 1, "shade", "dark"),
			),
		}

		for _, scenario := range scenarios {
			t.Run(scenario.description, func(t *testing.T) {
				modExpr, err := queryexpr.Parse(scenario.expression)
				assert.NoError(t, err)

				modExpr = modExpr.WithNameParams(scenario.placeholderNames).WithValueParams(scenario.placeholderValues)

				plan, err := modExpr.Plan(tableInfo)
				assert.NoError(t, err)

				assert.True(t, plan.CanQuery)
				assert.Equal(t, scenario.indexName, plan.IndexName)
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
			scanCase("when request sk starts with something", `sk^="something"`, `begins_with (#0, :0)`,
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
			scanCase("is true", `bool = true`, `#0 = :0`,
				exprName(0, "bool"),
				exprValueIsBool(0, true),
			),
			scanCase("is false", `bool = false`, `#0 = :0`,
				exprName(0, "bool"),
				exprValueIsBool(0, false),
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

			scanCase("with between", `pk between "a" and "z"`,
				`#0 BETWEEN :0 AND :1`,
				exprName(0, "pk"),
				exprValueIsString(0, "a"),
				exprValueIsString(1, "z"),
			),

			scanCase("with in", `pk in ("alpha", "bravo", "charlie")`,
				`#0 IN (:0, :1, :2)`,
				exprName(0, "pk"),
				exprValueIsString(0, "alpha"),
				exprValueIsString(1, "bravo"),
				exprValueIsString(2, "charlie"),
			),
			scanCase("with not in", `pk not in ("alpha", "bravo", "charlie")`,
				`NOT (#0 IN (:0, :1, :2))`,
				exprName(0, "pk"),
				exprValueIsString(0, "alpha"),
				exprValueIsString(1, "bravo"),
				exprValueIsString(2, "charlie"),
			),
			scanCase("with in with single operand returning a sequence", `pk in range(1, 5)`,
				`#0 IN (:0, :1, :2, :3, :4)`,
				exprName(0, "pk"),
				exprValueIsNumber(0, "1"),
				exprValueIsNumber(1, "2"),
				exprValueIsNumber(2, "3"),
				exprValueIsNumber(3, "4"),
				exprValueIsNumber(4, "5"),
			),
			scanCase("with in with single operand not returning a literal", `"foobar" in pk`,
				`contains (#0, :0)`,
				exprNameIsString(0, 0, "pk", "foobar"),
			),
			// TODO: in > 100 items ==> items OR items

			scanCase("with is S", `pk is "S"`,
				`attribute_type (#0, :0)`,
				exprNameIsString(0, 0, "pk", "S"),
			),
			scanCase("with is N", `pk is "N"`,
				`attribute_type (#0, :0)`,
				exprNameIsString(0, 0, "pk", "N"),
			),
			scanCase("with is not N", `pk is not "SS"`,
				`NOT (attribute_type (#0, :0))`,
				exprNameIsString(0, 0, "pk", "SS"),
			),
			scanCase("with is any", `pk is "any"`,
				`attribute_exists (#0)`,
				exprName(0, "pk"),
			),
			scanCase("with is not any", `pk is not "any"`,
				`attribute_not_exists (#0)`,
				exprName(0, "pk"),
			),

			scanCase("the size function as a right-side operand", `ln=size(pk)`,
				`#0 = size (#1)`,
				exprName(0, "ln"),
				exprName(1, "pk"),
			),
			scanCase("the size function as a left-side operand #1", `size(pk) = 123`,
				`size (#0) = :0`,
				exprNameIsNumber(0, 0, "pk", "123"),
			),
			scanCase("the size function as a left-side operand #2", `size(pk) > 123`,
				`size (#0) > :0`,
				exprNameIsNumber(0, 0, "pk", "123"),
			),
			scanCase("the size function on both sizes",
				`size(pk) != size(sk) and size(pk) > size(third) and size(pk) = 131`,
				`(size (#0) <> size (#1)) AND (size (#0) > size (#2)) AND (size (#0) = :0)`,
				exprName(0, "pk"),
				exprName(1, "sk"),
				exprName(2, "third"),
				exprValueIsNumber(0, "131"),
			),

			// Sub refs
			scanCase("with index", `this[2] = "something"`, `#0[2] = :0`,
				exprName(0, "this"),
				exprValueIsString(0, "something"),
			),
			scanCase("with the dot", `this.value = "something"`, `#0 = :0`,
				exprName(0, "this.value"),
				exprValueIsString(0, "something"),
			),
			/*
				scanCase("with multiple indices", `this[2][3] = "something"`, `#0[2][3] = :0`,
					exprName(0, "this"),
					exprValueIsString(0, "something"),
				),
				scanCase("with multiple indices with paren", `((this[2])[3])[4] = "something"`, `#0[2][3][4] = :0`,
					exprName(0, "this"),
					exprValueIsString(0, "something"),
				),
			*/
			scanCase("with multiple dots", `this.that.other.value = "else"`, `#0 = :0`,
				exprName(0, "this.that.other.value"),
				exprValueIsString(0, "else"),
			),

			// TODO: the contains function

			// Placeholders
			scanCase("with placeholders",
				`:partition=$valuePrefix or :sort=$valueAnother`,
				`(#0 = :0) OR (#1 = :1)`,
				placeholderNames(map[string]string{
					"partition": "pk",
					"sort":      "sk",
				}),
				placeholderValues(map[string]types.AttributeValue{
					"valuePrefix":  &types.AttributeValueMemberS{Value: "prefix"},
					"valueAnother": &types.AttributeValueMemberS{Value: "another"},
				}),
				exprNameIsString(0, 0, "pk", "prefix"),
				exprNameIsString(1, 1, "sk", "another"),
			),
		}

		for _, scenario := range scenarios {
			t.Run(scenario.description, func(t *testing.T) {
				modExpr, err := queryexpr.Parse(scenario.expression)
				assert.NoError(t, err)

				modExpr = modExpr.WithNameParams(scenario.placeholderNames).WithValueParams(scenario.placeholderValues)

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

	t.Run("with index clash", func(t *testing.T) {
		t.Run("should return error if attempt to run query with two indices that can be chosen", func(t *testing.T) {
			modExpr, err := queryexpr.Parse(`apples="this"`)
			assert.NoError(t, err)

			_, err = modExpr.Plan(tableInfo)
			assert.Error(t, err)
		})

		t.Run("should run as scan if explicitly forced to", func(t *testing.T) {
			modExpr, err := queryexpr.Parse(`apples="this" using scan`)
			assert.NoError(t, err)

			plan, err := modExpr.Plan(tableInfo)
			assert.NoError(t, err)
			assert.False(t, plan.CanQuery)
		})

		t.Run("should run as query with the 'with-apples' index", func(t *testing.T) {
			modExpr, err := queryexpr.Parse(`apples="this" using index("with-apples")`)
			assert.NoError(t, err)

			plan, err := modExpr.Plan(tableInfo)
			assert.NoError(t, err)
			assert.True(t, plan.CanQuery)
			assert.Equal(t, "with-apples", plan.IndexName)
		})

		t.Run("should run as query with the 'with-apples-and-oranges' index", func(t *testing.T) {
			modExpr, err := queryexpr.Parse(`apples="this" using index("with-apples-and-oranges")`)
			assert.NoError(t, err)

			plan, err := modExpr.Plan(tableInfo)
			assert.NoError(t, err)
			assert.True(t, plan.CanQuery)
			assert.Equal(t, "with-apples-and-oranges", plan.IndexName)
		})

		t.Run("should return error if the chosen index can't be used", func(t *testing.T) {
			modExpr, err := queryexpr.Parse(`apples="this" using index("with-missing")`)
			assert.NoError(t, err)

			_, err = modExpr.Plan(tableInfo)
			assert.Error(t, err)
		})
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
			"prime": &types.AttributeValueMemberL{
				Value: []types.AttributeValue{
					&types.AttributeValueMemberN{Value: "2"},
					&types.AttributeValueMemberN{Value: "3"},
					&types.AttributeValueMemberN{Value: "5"},
					&types.AttributeValueMemberN{Value: "7"},
				},
			},
			"three": &types.AttributeValueMemberN{Value: "3"},
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
			{expr: `missing`, expected: nil},

			// Equality with literal
			{expr: `alpha="alpha"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha!="not alpha"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `bravo=123`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `charlie.tree="green"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha^="al"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha="foobar"`, expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: `alpha^="need-something"`, expected: &types.AttributeValueMemberBOOL{Value: false}},

			// Comparison
			{expr: "three > 4", expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: "three >= 4", expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: "three < 4", expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: "three <= 4", expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: "three > 3", expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: "three >= 3", expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: "three < 3", expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: "three <= 3", expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: "three > 2", expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: "three >= 2", expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: "three < 2", expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: "three <= 2", expected: &types.AttributeValueMemberBOOL{Value: false}},

			// In
			{expr: "three in (2, 3, 4, 5)", expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: "three in (20, 30, 40)", expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: `alpha in ("alpha", "beta", "gamma", "delta")`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha in ("ey", "be", "see")`, expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: `three in prime`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `1 in prime`, expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: `"door" in charlie`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `"sky" in charlie`, expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: `"al" in alpha`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `"cent" in "percentage"`, expected: &types.AttributeValueMemberBOOL{Value: true}},

			// Is
			{expr: `alpha is "S"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `alpha is not "N"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `three is "N"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `three is not "S"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `(three = 3) is "BOOL"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `prime is "L"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `charlie is "M"`, expected: &types.AttributeValueMemberBOOL{Value: true}},

			{expr: `alpha is "any"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `three is "any"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `(three = 3) is "any"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `charlie is "any"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `prime is "any"`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `undef is not "any"`, expected: &types.AttributeValueMemberBOOL{Value: true}},

			// Size
			{expr: `size(alpha)`, expected: &types.AttributeValueMemberN{Value: "5"}},
			{expr: `size("This is a test")`, expected: &types.AttributeValueMemberN{Value: "14"}},
			{expr: `size(charlie)`, expected: &types.AttributeValueMemberN{Value: "2"}},
			{expr: `size(prime)`, expected: &types.AttributeValueMemberN{Value: "4"}},

			// Dot values
			{expr: `charlie.door`, expected: &types.AttributeValueMemberS{Value: "red"}},
			{expr: `(charlie).door`, expected: &types.AttributeValueMemberS{Value: "red"}},
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

			// Bool negation
			{expr: `not alpha="alpha"`, expected: &types.AttributeValueMemberBOOL{Value: false}},
			{expr: `not alpha!="alpha"`, expected: &types.AttributeValueMemberBOOL{Value: true}},

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

	t.Run("functions", func(t *testing.T) {
		timeNow := time.Now()

		scenarios := []struct {
			expr     string
			expected types.AttributeValue
		}{
			// _x_now() -- unreleased version of now
			{expr: `_x_now()`, expected: &types.AttributeValueMemberN{Value: fmt.Sprint(timeNow.Unix())}},
		}
		for _, scenario := range scenarios {
			t.Run(scenario.expr, func(t *testing.T) {
				modExpr, err := queryexpr.Parse(scenario.expr)
				assert.NoError(t, err)

				res, err := modExpr.WithTestTimeSource(timeNow).EvalItem(item)
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

	t.Run("name and value placeholders", func(t *testing.T) {
		scenarios := []struct {
			expr     string
			expected types.AttributeValue
		}{
			{expr: `alpha = $a`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `:theBName = 123`, expected: &types.AttributeValueMemberBOOL{Value: true}},
			{expr: `:theCMap.door`, expected: &types.AttributeValueMemberS{Value: "red"}},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.expr, func(t *testing.T) {
				modExpr, err := queryexpr.Parse(scenario.expr)
				assert.NoError(t, err)

				modExpr = modExpr.WithValueParams(map[string]types.AttributeValue{
					"a": &types.AttributeValueMemberS{Value: "alpha"},
				}).WithNameParams(map[string]string{
					"theBName": "bravo",
					"theCMap":  "charlie",
				})

				res, err := modExpr.EvalItem(item)
				assert.NoError(t, err)

				assert.Equal(t, scenario.expected, res)
			})
		}
	})
}

func TestQueryExpr_SetEvalItem(t *testing.T) {
	var templateItem = func() models.Item {
		return models.Item{
			"alpha": &types.AttributeValueMemberS{Value: "alpha"},
			"bravo": &types.AttributeValueMemberN{Value: "123"},
			"charlie": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"door": &types.AttributeValueMemberS{Value: "red"},
					"tree": &types.AttributeValueMemberS{Value: "green"},
				},
			},
			"prime": &types.AttributeValueMemberL{
				Value: []types.AttributeValue{
					&types.AttributeValueMemberN{Value: "2"},
					&types.AttributeValueMemberN{Value: "3"},
					&types.AttributeValueMemberN{Value: "5"},
					&types.AttributeValueMemberN{Value: "7"},
				},
			},
			"three": &types.AttributeValueMemberN{Value: "3"},
		}
	}

	t.Run("simple values", func(t *testing.T) {
		item := templateItem()

		modExpr, err := queryexpr.Parse("alpha")
		assert.NoError(t, err)
		assert.True(t, modExpr.IsModifiablePath(item))

		err = modExpr.SetEvalItem(item, &types.AttributeValueMemberS{Value: "not alpha"})
		assert.NoError(t, err)
		assert.Equal(t, "not alpha", item["alpha"].(*types.AttributeValueMemberS).Value)
	})

	t.Run("dot values", func(t *testing.T) {
		item := templateItem()

		modExpr, err := queryexpr.Parse("charlie.tree")
		assert.NoError(t, err)
		assert.True(t, modExpr.IsModifiablePath(item))

		err = modExpr.SetEvalItem(item, &types.AttributeValueMemberS{Value: "Birch"})
		assert.NoError(t, err)
		assert.Equal(t, "Birch", item["charlie"].(*types.AttributeValueMemberM).Value["tree"].(*types.AttributeValueMemberS).Value)
	})
}

func TestQueryExpr_DeleteAttribute(t *testing.T) {
	var templateItem = func() models.Item {
		return models.Item{
			"alpha": &types.AttributeValueMemberS{Value: "alpha"},
			"bravo": &types.AttributeValueMemberN{Value: "123"},
			"charlie": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"door": &types.AttributeValueMemberS{Value: "red"},
					"tree": &types.AttributeValueMemberS{Value: "green"},
				},
			},
			"prime": &types.AttributeValueMemberL{
				Value: []types.AttributeValue{
					&types.AttributeValueMemberN{Value: "2"},
					&types.AttributeValueMemberN{Value: "3"},
					&types.AttributeValueMemberN{Value: "5"},
					&types.AttributeValueMemberN{Value: "7"},
				},
			},
			"three": &types.AttributeValueMemberN{Value: "3"},
		}
	}

	t.Run("simple values", func(t *testing.T) {
		item := templateItem()

		modExpr, err := queryexpr.Parse("alpha")
		assert.NoError(t, err)

		err = modExpr.DeleteAttribute(item)
		assert.NoError(t, err)

		_, hasKey := item["alpha"]
		assert.False(t, hasKey)
	})

	t.Run("placeholder values", func(t *testing.T) {
		item := templateItem()

		modExpr, err := queryexpr.Parse(":a")
		assert.NoError(t, err)

		modExpr = modExpr.WithNameParams(map[string]string{"a": "alpha"})

		err = modExpr.DeleteAttribute(item)
		assert.NoError(t, err)

		_, hasKey := item["alpha"]
		assert.False(t, hasKey)
	})

	t.Run("dot values", func(t *testing.T) {
		item := templateItem()

		modExpr, err := queryexpr.Parse("charlie.tree")
		assert.NoError(t, err)

		err = modExpr.DeleteAttribute(item)
		assert.NoError(t, err)

		_, hasKey := item["charlie"].(*types.AttributeValueMemberM).Value["tree"]
		assert.False(t, hasKey)
	})

	t.Run("dot values with placeholders", func(t *testing.T) {
		item := templateItem()

		modExpr, err := queryexpr.Parse(":c.tree")
		assert.NoError(t, err)

		modExpr = modExpr.WithNameParams(map[string]string{"c": "charlie"})

		err = modExpr.DeleteAttribute(item)
		assert.NoError(t, err)

		_, hasKey := item["charlie"].(*types.AttributeValueMemberM).Value["tree"]
		assert.False(t, hasKey)
	})

	//t.Run("dot values with multiple placeholders", func(t *testing.T) {
	//	item := templateItem()
	//
	//	modExpr, err := queryexpr.Parse(":c.:t")
	//	assert.NoError(t, err)
	//
	//	modExpr = modExpr.WithNameParams(map[string]string{
	//		"c": "charlie",
	//		"t": "tree",
	//	})
	//
	//	err = modExpr.DeleteAttribute(item)
	//	assert.NoError(t, err)
	//
	//	_, hasKey := item["charlie"].(*types.AttributeValueMemberM).Value["tree"]
	//	assert.False(t, hasKey)
	//})
}

func TestQueryExpr_SerializeTo(t *testing.T) {
	t.Run("should be able to serialized and deseralize the parsed expression", func(t *testing.T) {
		exprStr := `something = $value and :placeholder = "something else" and thirdThing in (1,2,3)`

		bts := new(bytes.Buffer)

		modExpr, err := queryexpr.Parse(exprStr)
		assert.NoError(t, err)

		modExpr = modExpr.WithNameParams(map[string]string{
			"placeholder": "some name",
		}).WithValueParams(map[string]types.AttributeValue{
			"value":           &types.AttributeValueMemberS{Value: "some value"},
			"num":             &types.AttributeValueMemberN{Value: "12345"},
			"veryLargeNumber": &types.AttributeValueMemberN{Value: "123456789012345678901234567890"},
			"numberSet":       &types.AttributeValueMemberNS{Value: []string{"123", "234", "345"}},
			"bool":            &types.AttributeValueMemberBOOL{Value: true},
			"list": &types.AttributeValueMemberL{Value: []types.AttributeValue{
				&types.AttributeValueMemberN{Value: "1"},
				&types.AttributeValueMemberN{Value: "2"},
				&types.AttributeValueMemberN{Value: "3"},
			}},
			"dict": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"alpha":   &types.AttributeValueMemberS{Value: "apple"},
					"bravo":   &types.AttributeValueMemberS{Value: "banana"},
					"charlie": &types.AttributeValueMemberS{Value: "cherry"},
				},
			},
		})

		assert.NoError(t, modExpr.SerializeTo(bts))

		newExpr, err := queryexpr.DeserializeFrom(bts)
		assert.NoError(t, err)
		assert.Equal(t, modExpr.String(), newExpr.String())

		name, hasName := newExpr.NameParam("placeholder")
		assert.Equal(t, "some name", name)
		assert.True(t, hasName)

		assert.Equal(t, "some value", newExpr.ValueParamOrNil("value").(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "12345", newExpr.ValueParamOrNil("num").(*types.AttributeValueMemberN).Value)
		assert.Equal(t, "123456789012345678901234567890", newExpr.ValueParamOrNil("veryLargeNumber").(*types.AttributeValueMemberN).Value)
		assert.Equal(t, []string{"123", "234", "345"}, newExpr.ValueParamOrNil("numberSet").(*types.AttributeValueMemberNS).Value)
		assert.Equal(t, true, newExpr.ValueParamOrNil("bool").(*types.AttributeValueMemberBOOL).Value)
		assert.Equal(t, "1", newExpr.ValueParamOrNil("list").(*types.AttributeValueMemberL).Value[0].(*types.AttributeValueMemberN).Value)
		assert.Equal(t, "2", newExpr.ValueParamOrNil("list").(*types.AttributeValueMemberL).Value[1].(*types.AttributeValueMemberN).Value)
		assert.Equal(t, "3", newExpr.ValueParamOrNil("list").(*types.AttributeValueMemberL).Value[2].(*types.AttributeValueMemberN).Value)
		assert.Equal(t, "apple", newExpr.ValueParamOrNil("dict").(*types.AttributeValueMemberM).Value["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "banana", newExpr.ValueParamOrNil("dict").(*types.AttributeValueMemberM).Value["bravo"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "cherry", newExpr.ValueParamOrNil("dict").(*types.AttributeValueMemberM).Value["charlie"].(*types.AttributeValueMemberS).Value)
	})
}

func TestQueryExpr_Equals(t *testing.T) {
	t.Run("should perform equals correctly", func(t *testing.T) {
		exprStr := `something = $value and :placeholder = "something else" and thirdThing in (1,2,3)`

		modExpr, _ := queryexpr.Parse(exprStr)
		modExpr = modExpr.WithNameParams(map[string]string{
			"placeholder": "some name",
			"another":     "name",
			"more":        "names",
		}).WithValueParams(map[string]types.AttributeValue{
			"value":           &types.AttributeValueMemberS{Value: "some value"},
			"num":             &types.AttributeValueMemberN{Value: "12345"},
			"veryLargeNumber": &types.AttributeValueMemberN{Value: "123456789012345678901234567890"},
			"numberSet":       &types.AttributeValueMemberNS{Value: []string{"123", "234", "345"}},
			"bool":            &types.AttributeValueMemberBOOL{Value: true},
			"list": &types.AttributeValueMemberL{Value: []types.AttributeValue{
				&types.AttributeValueMemberN{Value: "1"},
				&types.AttributeValueMemberN{Value: "2"},
				&types.AttributeValueMemberN{Value: "3"},
			}},
			"dict": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"alpha":   &types.AttributeValueMemberS{Value: "apple"},
					"bravo":   &types.AttributeValueMemberS{Value: "banana"},
					"charlie": &types.AttributeValueMemberS{Value: "cherry"},
				},
			},
		})

		differentExpr, _ := queryexpr.Parse(`abc = :bla`)
		differentExpr = modExpr.WithNameParams(map[string]string{
			"fla": "some name",
		}).WithValueParams(map[string]types.AttributeValue{
			"value": &types.AttributeValueMemberS{Value: "some value"},
		})

		bts1, err := modExpr.SerializeToBytes()
		assert.NoError(t, err)

		expr2, err := queryexpr.DeserializeFrom(bytes.NewReader(bts1))
		assert.NoError(t, err)

		bts2, err := expr2.SerializeToBytes()
		assert.NoError(t, err)

		expr3, err := queryexpr.DeserializeFrom(bytes.NewReader(bts2))
		assert.NoError(t, err)

		_, err = expr3.SerializeToBytes()
		assert.NoError(t, err)

		var nilQE *queryexpr.QueryExpr
		assert.True(t, nilQE.Equal(nil))
		assert.True(t, modExpr.Equal(expr2))
		assert.True(t, expr2.Equal(expr3))
		assert.True(t, expr3.Equal(modExpr))

		assert.False(t, nilQE.Equal(differentExpr))
		assert.False(t, modExpr.Equal(differentExpr))
		assert.False(t, expr2.Equal(differentExpr))
		assert.False(t, expr3.Equal(differentExpr))

		assert.Equal(t, uint64(0), nilQE.HashCode())
		assert.Equal(t, modExpr.HashCode(), expr2.HashCode())
		assert.Equal(t, expr2.HashCode(), expr3.HashCode())
		assert.Equal(t, expr3.HashCode(), modExpr.HashCode())

		assert.NotEqual(t, differentExpr.HashCode(), nilQE.HashCode())
		assert.NotEqual(t, differentExpr.HashCode(), expr2.HashCode())
		assert.NotEqual(t, differentExpr.HashCode(), expr3.HashCode())
		assert.NotEqual(t, differentExpr.HashCode(), modExpr.HashCode())
	})
}

type scanScenario struct {
	description       string
	expression        string
	expectedFilter    string
	indexName         string
	expectedNames     map[string]string
	expectedValues    map[string]types.AttributeValue
	placeholderNames  map[string]string
	placeholderValues map[string]types.AttributeValue
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

func indexName(indexName string) func(ss *scanScenario) {
	return func(ss *scanScenario) {
		ss.indexName = indexName
	}
}

func placeholderNames(placeholderNames map[string]string) func(ss *scanScenario) {
	return func(ss *scanScenario) {
		ss.placeholderNames = placeholderNames
	}
}

func placeholderValues(placeholderValues map[string]types.AttributeValue) func(ss *scanScenario) {
	return func(ss *scanScenario) {
		ss.placeholderValues = placeholderValues
	}
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

func exprValueIsNumber(valIdx int, expected string) func(ss *scanScenario) {
	return func(ss *scanScenario) {
		ss.expectedValues[fmt.Sprintf(":%d", valIdx)] = &types.AttributeValueMemberN{Value: expected}
	}
}

func exprValueIsBool(valIdx int, expected bool) func(ss *scanScenario) {
	return func(ss *scanScenario) {
		ss.expectedValues[fmt.Sprintf(":%d", valIdx)] = &types.AttributeValueMemberBOOL{Value: expected}
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
