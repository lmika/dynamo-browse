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
		scenarios := []struct {
			desc string
			code string
		}{
			{
				desc: "single function, table name match",
				code: `
					ext.related_items("test-table", func(item) {
						print("Hello")
						return [
							{"label": "Customer", "query": "pk=$foo", "args": {"foo": "foo"}},
							{"label": "Payment", "query": "fla=$daa", "args": {"daa": "Hello"}},
						]
					})	
				`,
			},
			{
				desc: "single function, table prefix match",
				code: `
					ext.related_items("test-*", func(item) {
						print("Hello")
						return [
							{"label": "Customer", "query": "pk=$foo", "args": {"foo": "foo"}},
							{"label": "Payment", "query": "fla=$daa", "args": {"daa": "Hello"}},
						]
					})	
				`,
			},
			{
				desc: "multi function, table name match",
				code: `
					ext.related_items("test-table", func(item) {
						print("Hello")
						return [
							{"label": "Customer", "query": "pk=$foo", "args": {"foo": "foo"}},
						]
					})

					ext.related_items("test-table", func(item) {
						return [
							{"label": "Payment", "query": "fla=$daa", "args": {"daa": "Hello"}},
						]
					})	
				`,
			},
			{
				desc: "multi function, table name prefix",
				code: `
					ext.related_items("test-*", func(item) {
						print("Hello")
						return [
							{"label": "Customer", "query": "pk=$foo", "args": {"foo": "foo"}},
						]
					})

					ext.related_items("test-*", func(item) {
						return [
							{"label": "Payment", "query": "fla=$daa", "args": {"daa": "Hello"}},
						]
					})	
				`,
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.desc, func(t *testing.T) {
				// Load the script
				srv := scriptmanager.New(scriptmanager.WithFS(testScriptFile(t, "test.tm", scenario.code)))

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
				assert.Equal(t, "foo", relItems[0].Query.ValueParamOrNil("foo").(*types.AttributeValueMemberS).Value)

				assert.Equal(t, "Payment", relItems[1].Name)
				assert.Equal(t, "fla=$daa", relItems[1].Query.String())
				assert.Equal(t, "Hello", relItems[1].Query.ValueParamOrNil("daa").(*types.AttributeValueMemberS).Value)
			})
		}
	})

	t.Run("should support rel_items with on select", func(t *testing.T) {
		// Load the script
		srv := scriptmanager.New(scriptmanager.WithFS(testScriptFile(t, "test.tm", `
			ext.related_items("test-table", func(item) {
				print("Hello")
				return [
					{"label": "Customer", "on_select": func() {
						print("Selected")
					}},
				]
			})	
		`)))

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
		assert.Len(t, relItems, 1)

		assert.Equal(t, "Customer", relItems[0].Name)
		assert.NoError(t, relItems[0].OnSelect())
	})
}
