package modexpr_test

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/lmika/awstools/internal/dynamo-browse/models/modexpr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModExpr_Patch(t *testing.T) {
	t.Run("patch with new attributes", func(t *testing.T) {
		modExpr, err := modexpr.Parse(`alpha="new value", beta="another new value"`)
		assert.NoError(t, err)

		oldItem := models.Item{}
		newItem, err := modExpr.Patch(oldItem)
		assert.NoError(t, err)

		assert.Equal(t, "new value", newItem["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "another new value", newItem["beta"].(*types.AttributeValueMemberS).Value)
	})

	t.Run("patch with existing attributes", func(t *testing.T) {
		modExpr, err := modexpr.Parse(`alpha="new value", beta="another new value"`)
		assert.NoError(t, err)

		oldItem := models.Item{
			"old": &types.AttributeValueMemberS{Value: "before"},
			"beta": &types.AttributeValueMemberS{Value: "before beta"},
		}
		newItem, err := modExpr.Patch(oldItem)
		assert.NoError(t, err)

		assert.Equal(t, "before", newItem["old"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "new value", newItem["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "another new value", newItem["beta"].(*types.AttributeValueMemberS).Value)
	})
}
