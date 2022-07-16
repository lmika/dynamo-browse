package dynamo_test

import (
	"context"
	"fmt"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/test/testdynamo"
	"github.com/stretchr/testify/assert"
)

func TestProvider_ScanItems(t *testing.T) {
	tableName := "test-table"

	client := testdynamo.SetupTestTable(t, testData)
	provider := dynamo.NewProvider(client)

	t.Run("should return scanned items from the table", func(t *testing.T) {
		ctx := context.Background()

		items, err := provider.ScanItems(ctx, tableName, nil, 100)
		assert.NoError(t, err)
		assert.Len(t, items, 3)

		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[0].Data[0]))
		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[0].Data[1]))
		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[0].Data[2]))
	})

	t.Run("should return error if table name does not exist", func(t *testing.T) {
		ctx := context.Background()

		items, err := provider.ScanItems(ctx, "does-not-exist", nil, 100)
		assert.Error(t, err)
		assert.Nil(t, items)
	})
}

func TestProvider_PutItems(t *testing.T) {
	tableName := "test-table"

	scenarios := []struct {
		maxItems int
	}{
		{maxItems: 3},
		{maxItems: 13},
		{maxItems: 25},
		{maxItems: 48},
		{maxItems: 73},
		{maxItems: 103},
		{maxItems: 291},
	}
	for _, scenario := range scenarios {
		t.Run(fmt.Sprintf("should put items in batches: size %v", scenario.maxItems), func(t *testing.T) {
			ctx := context.Background()

			client := testdynamo.SetupTestTable(t, []testdynamo.TestData{
				{
					TableName: tableName,
				},
			})

			provider := dynamo.NewProvider(client)

			items := make([]models.Item, scenario.maxItems)
			for i := 0; i < scenario.maxItems; i++ {
				items[i] = models.Item{
					"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("K#%v", i)},
					"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("K#%v", i)},
					"a":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%v", i)},
				}
			}

			// Write the data
			err := provider.PutItems(ctx, tableName, items)
			assert.NoError(t, err)

			// Verify the data
			readItems, err := provider.ScanItems(ctx, tableName, nil, scenario.maxItems+5)
			assert.NoError(t, err)
			assert.Len(t, readItems, scenario.maxItems)

			for i := 0; i < scenario.maxItems; i++ {
				assert.Contains(t, readItems, items[i])
			}
		})
	}
}

func TestProvider_DeleteItem(t *testing.T) {
	tableName := "test-table"

	t.Run("should delete item if exists in table", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)
		provider := dynamo.NewProvider(client)

		ctx := context.Background()

		err := provider.DeleteItem(ctx, tableName, map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "abc"},
			"sk": &types.AttributeValueMemberS{Value: "222"},
		})

		items, err := provider.ScanItems(ctx, tableName, nil, 100)
		assert.NoError(t, err)
		assert.Len(t, items, 2)

		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[0].Data[0]))
		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[0].Data[2]))
		assert.NotContains(t, items, testdynamo.TestRecordAsItem(t, testData[0].Data[1]))

	})

	t.Run("should do nothing if key does not exist", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)
		provider := dynamo.NewProvider(client)

		ctx := context.Background()

		err := provider.DeleteItem(ctx, tableName, map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "zyx"},
			"sk": &types.AttributeValueMemberS{Value: "999"},
		})

		items, err := provider.ScanItems(ctx, tableName, nil, 100)
		assert.NoError(t, err)
		assert.Len(t, items, 3)

		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[0].Data[0]))
		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[0].Data[1]))
		assert.Contains(t, items, testdynamo.TestRecordAsItem(t, testData[0].Data[2]))
	})

	t.Run("should return error if table name does not exist", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)
		provider := dynamo.NewProvider(client)

		ctx := context.Background()

		items, err := provider.ScanItems(ctx, "does-not-exist", nil, 100)
		assert.Error(t, err)
		assert.Nil(t, items)
	})
}

var testData = []testdynamo.TestData{
	{
		TableName: "test-table",
		Data: []map[string]interface{}{
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
		},
	},
}
