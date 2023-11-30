package scriptmanager_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/scriptmanager"
	"github.com/stretchr/testify/assert"
)

func TestExtModule_RelatedItems(t *testing.T) {
	t.Run("should register a function which will return related items for an item", func(t *testing.T) {
		// Load the script
		testFS := testScriptFile(t, "test.tm", `
			ext.related_items("test-table", func(item) {
				print("Hello")
				return [
					{"label": "Customer", "query": "pk=$foo", "args": {"foo": "foo"}},
					{"label": "Payment", "query": "fla=$daa", "args": {"daa": "Hello"}},
				]
			})
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))

		ctx := context.Background()
		plugin, err := srv.LoadScript(ctx, "test.tm")
		assert.NoError(t, err)
		assert.NotNil(t, plugin)

		// Get related items of result set
		rs := &models.ResultSet{
			TableInfo: &models.TableInfo{
				Name: "test-table",
			},
		}
		rs.SetItems([]models.Item{
			{"pk": &types.AttributeValueMemberS{Value: "abc"}},
			{"pk": &types.AttributeValueMemberS{Value: "1232"}},
		})

		relItems, err := srv.RelatedItemOfItem(context.Background(), rs, 0)
		assert.NoError(t, err)
		assert.Len(t, relItems, 2)

		assert.Equal(t, "Customer", relItems[0].Name)
		assert.Equal(t, "pk=$foo", relItems[0].Query.String())

		assert.Equal(t, "Payment", relItems[1].Name)
		assert.Equal(t, "fla=$daa", relItems[1].Query.String())
	})
}
