package scriptmanager_test

import (
	"context"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/scriptmanager"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtModule_RelatedItems(t *testing.T) {
	t.Run("should register a function which will return related items for an item", func(t *testing.T) {
		testFS := testScriptFile(t, "test.tm", `
			ext.related_items("test-table", func(item) {
				return [
					{"label": "Customer", "query": "pk=$foo", "args": {"foo": "foo"}},
					{"label": "Payment", "query": "fla=$daa", "args": {"daa": "Hello"}}
				]
			})
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))

		ctx := context.Background()
		plugin, err := srv.LoadScript(ctx, "test.tm")
		assert.NoError(t, err)
		assert.NotNil(t, plugin)
	})
}
