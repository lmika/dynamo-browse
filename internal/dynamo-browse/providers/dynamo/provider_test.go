package dynamo_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/test/testdynamo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProvider_ScanItems(t *testing.T) {
	tableName := "provider-scanimages-test-table"

	client, cleanupFn := testdynamo.SetupTestTable(t, tableName, testData)
	defer cleanupFn()
	provider := dynamo.NewProvider(client)

	t.Run("should return scanned items from the table", func(t *testing.T) {
		ctx := context.Background()

		items, err := provider.ScanItems(ctx, tableName)
		assert.NoError(t, err)
		assert.Len(t, items, 3)

		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[0]))
		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[1]))
		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[2]))
	})

	t.Run("should return error if table name does not exist", func(t *testing.T) {
		ctx := context.Background()

		items, err := provider.ScanItems(ctx, "does-not-exist")
		assert.Error(t, err)
		assert.Nil(t, items)
	})
}

func TestProvider_DeleteItem(t *testing.T) {
	tableName := "provider-deleteitem-test-table"

	t.Run("should delete item if exists in table", func(t *testing.T) {
		client, cleanupFn := testdynamo.SetupTestTable(t, tableName, testData)
		defer cleanupFn()
		provider := dynamo.NewProvider(client)

		ctx := context.Background()

		err := provider.DeleteItem(ctx, tableName, map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "abc"},
			"sk": &types.AttributeValueMemberS{Value: "222"},
		})

		items, err := provider.ScanItems(ctx, tableName)
		assert.NoError(t, err)
		assert.Len(t, items, 2)

		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[0]))
		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[2]))
		assert.NotContains(t, items, testdynamo.TestRecordAsItem(t, testData[1]))

	})

	t.Run("should do nothing if key does not exist", func(t *testing.T) {
		client, cleanupFn := testdynamo.SetupTestTable(t, tableName, testData)
		defer cleanupFn()
		provider := dynamo.NewProvider(client)

		ctx := context.Background()

		err := provider.DeleteItem(ctx, tableName, map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "zyx"},
			"sk": &types.AttributeValueMemberS{Value: "999"},
		})

		items, err := provider.ScanItems(ctx, tableName)
		assert.NoError(t, err)
		assert.Len(t, items, 3)

		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[0]))
		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[1]))
		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[2]))
	})

	t.Run("should return error if table name does not exist", func(t *testing.T) {
		client, cleanupFn := testdynamo.SetupTestTable(t, tableName, testData)
		defer cleanupFn()
		provider := dynamo.NewProvider(client)

		ctx := context.Background()

		items, err := provider.ScanItems(ctx, "does-not-exist")
		assert.Error(t, err)
		assert.Nil(t, items)
	})
}

var testData = testdynamo.TestData{
	{
		"pk":    "abc",
		"sk":    "111",
		"alpha": "This is some value",
	},
	{
		"pk":    "abc",
		"sk":    "222",
		"alpha": "This is another some value",
		"beta":  1231,
	},
	{
		"pk":    "bbb",
		"sk":    "131",
		"beta":  2468,
		"gamma": "foobar",
	},
}
