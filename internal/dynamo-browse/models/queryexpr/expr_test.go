package queryexpr_test

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/awstools/internal/dynamo-browse/models/queryexpr"
	"testing"

	"github.com/lmika/awstools/internal/dynamo-browse/models"
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

		assert.False(t, plan.CanQuery)
		assert.Equal(t, "#0 = :0", aws.ToString(plan.Expression.Filter()))
		assert.Equal(t, "pk", plan.Expression.Names()["#0"])
		assert.Equal(t, "prefix", plan.Expression.Values()[":0"].(*types.AttributeValueMemberS).Value)
	})

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
