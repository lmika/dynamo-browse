package queryexpr_test

import (
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
		t.Run("perform query when request pk is fixed", func(t *testing.T) {
			modExpr, err := queryexpr.Parse(`pk="prefix"`)
			assert.NoError(t, err)

			plan, err := modExpr.Plan(tableInfo)
			assert.NoError(t, err)

			assert.True(t, plan.CanQuery)
			assert.Equal(t, "#0 = :0", aws.ToString(plan.Expression.KeyCondition()))
			assert.Equal(t, "pk", plan.Expression.Names()["#0"])
			assert.Equal(t, "prefix", plan.Expression.Values()[":0"].(*types.AttributeValueMemberS).Value)
		})

		t.Run("perform query when request pk and sk is fixed", func(t *testing.T) {
			modExpr, err := queryexpr.Parse(`pk="prefix" and sk="another"`)
			assert.NoError(t, err)

			plan, err := modExpr.Plan(tableInfo)
			assert.NoError(t, err)

			assert.True(t, plan.CanQuery)
			assert.Equal(t, "(#0 = :0) AND (#1 = :1)", aws.ToString(plan.Expression.KeyCondition()))
			assert.Equal(t, "pk", plan.Expression.Names()["#0"])
			assert.Equal(t, "sk", plan.Expression.Names()["#1"])
			assert.Equal(t, "prefix", plan.Expression.Values()[":0"].(*types.AttributeValueMemberS).Value)
			assert.Equal(t, "another", plan.Expression.Values()[":1"].(*types.AttributeValueMemberS).Value)
		})

		t.Run("perform query when request pk is equals and sk is prefix", func(t *testing.T) {
			scenarios := []struct {
				expr string
			}{
				{expr: `pk="prefix" and sk^="another"`},
				{expr: `sk^="another" and pk="prefix"`},
			}

			for _, scenario := range scenarios {
				t.Run(scenario.expr, func(t *testing.T) {
					modExpr, err := queryexpr.Parse(scenario.expr)
					assert.NoError(t, err)

					plan, err := modExpr.Plan(tableInfo)
					assert.NoError(t, err)

					assert.True(t, plan.CanQuery)
					assert.Equal(t, "(#0 = :0) AND (begins_with (#1, :1))", aws.ToString(plan.Expression.KeyCondition()))
					assert.Equal(t, "pk", plan.Expression.Names()["#0"])
					assert.Equal(t, "sk", plan.Expression.Names()["#1"])
					assert.Equal(t, "prefix", plan.Expression.Values()[":0"].(*types.AttributeValueMemberS).Value)
					assert.Equal(t, "another", plan.Expression.Values()[":1"].(*types.AttributeValueMemberS).Value)
				})
			}
		})
	})

	t.Run("as scans", func(t *testing.T) {
		t.Run("when request pk prefix", func(t *testing.T) {
			modExpr, err := queryexpr.Parse(`pk^="prefix"`) // TODO: fix this so that '^ =' is invalid
			assert.NoError(t, err)

			plan, err := modExpr.Plan(tableInfo)
			assert.NoError(t, err)

			assert.False(t, plan.CanQuery)
			assert.Equal(t, "begins_with (#0, :0)", aws.ToString(plan.Expression.Filter()))
			assert.Equal(t, "pk", plan.Expression.Names()["#0"])
			assert.Equal(t, "prefix", plan.Expression.Values()[":0"].(*types.AttributeValueMemberS).Value)
		})

		t.Run("when request sk equals something", func(t *testing.T) {
			modExpr, err := queryexpr.Parse(`sk="something"`)
			assert.NoError(t, err)

			plan, err := modExpr.Plan(tableInfo)
			assert.NoError(t, err)

			assert.False(t, plan.CanQuery)
			assert.Equal(t, "#0 = :0", aws.ToString(plan.Expression.Filter()))
			assert.Equal(t, "sk", plan.Expression.Names()["#0"])
			assert.Equal(t, "something", plan.Expression.Values()[":0"].(*types.AttributeValueMemberS).Value)
		})

		t.Run("with disjunctions", func(t *testing.T) {
			modExpr, err := queryexpr.Parse(`pk="prefix" or sk="another"`)
			assert.NoError(t, err)

			plan, err := modExpr.Plan(tableInfo)
			assert.NoError(t, err)

			assert.False(t, plan.CanQuery)
			assert.Equal(t, "(#0 = :0) OR (#1 = :1)", aws.ToString(plan.Expression.Filter()))
			assert.Equal(t, "pk", plan.Expression.Names()["#0"])
			assert.Equal(t, "sk", plan.Expression.Names()["#1"])
			assert.Equal(t, "prefix", plan.Expression.Values()[":0"].(*types.AttributeValueMemberS).Value)
			assert.Equal(t, "another", plan.Expression.Values()[":1"].(*types.AttributeValueMemberS).Value)
		})

		t.Run("with disjunctions if pk is present twice in expression", func(t *testing.T) {
			modExpr, err := queryexpr.Parse(`pk="prefix" and pk="another"`)
			assert.NoError(t, err)

			plan, err := modExpr.Plan(tableInfo)
			assert.NoError(t, err)

			assert.False(t, plan.CanQuery)
			assert.Equal(t, "(#0 = :0) AND (#0 = :1)", aws.ToString(plan.Expression.Filter()))
			assert.Equal(t, "pk", plan.Expression.Names()["#0"])
			assert.Equal(t, "prefix", plan.Expression.Values()[":0"].(*types.AttributeValueMemberS).Value)
			assert.Equal(t, "another", plan.Expression.Values()[":1"].(*types.AttributeValueMemberS).Value)
		})
	})

	t.Run("perform scan with disjunction", func(t *testing.T) {
		modExpr, err := queryexpr.Parse(`pk^="prefix" or pk="prefix"`) // TODO: fix this so that '^ =' is invalid
		assert.NoError(t, err)

		plan, err := modExpr.Plan(tableInfo)
		assert.NoError(t, err)

		assert.False(t, plan.CanQuery)
		assert.Equal(t, "(begins_with (#0, :0)) OR (#0 = :1)", aws.ToString(plan.Expression.Filter()))
		assert.Equal(t, "pk", plan.Expression.Names()["#0"])
		assert.Equal(t, "prefix", plan.Expression.Values()[":0"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "prefix", plan.Expression.Values()[":1"].(*types.AttributeValueMemberS).Value)
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

			// Dot values
			{expr: `charlie.door`, expected: &types.AttributeValueMemberS{Value: "red"}},
			{expr: `charlie.tree`, expected: &types.AttributeValueMemberS{Value: "green"}},
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
