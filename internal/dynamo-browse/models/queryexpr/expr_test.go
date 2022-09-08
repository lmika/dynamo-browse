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

	/*
		t.Run("perform query when request pk and sk are fixed", func(t *testing.T) {
			scenarios := []struct {
				expr              string
				expectedKeyFilter string
			}{
				{
					expr:              `pk="prefix" and sk="second"`,
					expectedKeyFilter: "#0 = :0 AND #1 = :1",
				},
				{
					expr:              `sk="prefix" and pk="second"`,
					expectedKeyFilter: "#0 = :0 AND #1 = :1",
				},
			}

			for _, scenario := range scenarios {
				t.Run(scenario.expr, func(t *testing.T) {
					modExpr, err := queryexpr.Parse(scenario.expr)
					assert.NoError(t, err)

					plan, err := modExpr.Plan(tableInfo)
					assert.NoError(t, err)

					assert.True(t, plan.CanQuery)
					assert.Equal(t, scenario.expectedKeyFilter, aws.ToString(plan.Expression.KeyCondition()))
					assert.Equal(t, "pk", plan.Expression.Names()["#0"])
					assert.Equal(t, "sk", plan.Expression.Names()["#1"])
					assert.Equal(t, "prefix", plan.Expression.Values()[":0"].(*types.AttributeValueMemberS).Value)
					assert.Equal(t, "second", plan.Expression.Values()[":1"].(*types.AttributeValueMemberS).Value)
				})
			}
		})
	*/

	t.Run("perform scan when request pk prefix", func(t *testing.T) {
		modExpr, err := queryexpr.Parse(`pk^="prefix"`) // TODO: fix this so that '^ =' is invalid
		assert.NoError(t, err)

		plan, err := modExpr.Plan(tableInfo)
		assert.NoError(t, err)

		assert.False(t, plan.CanQuery)
		assert.Equal(t, "begins_with (#0, :0)", aws.ToString(plan.Expression.Filter()))
		assert.Equal(t, "pk", plan.Expression.Names()["#0"])
		assert.Equal(t, "prefix", plan.Expression.Values()[":0"].(*types.AttributeValueMemberS).Value)
	})
}
