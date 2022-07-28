package models_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/stretchr/testify/assert"
)

func TestSort(t *testing.T) {
	t.Run("pk and sk are both strings", func(t *testing.T) {
		tableInfo := &models.TableInfo{Keys: models.KeyAttribute{PartitionKey: "pk", SortKey: "sk"}}

		items := make([]models.Item, len(testStringData))
		copy(items, testStringData)

		models.Sort(items, tableInfo)

		assert.Equal(t, items[0], testStringData[1])
		assert.Equal(t, items[1], testStringData[2])
		assert.Equal(t, items[2], testStringData[0])
	})

	t.Run("pk and sk are both numbers", func(t *testing.T) {
		tableInfo := &models.TableInfo{Keys: models.KeyAttribute{PartitionKey: "pk", SortKey: "sk"}}

		items := make([]models.Item, len(testNumberData))
		copy(items, testNumberData)

		models.Sort(items, tableInfo)

		assert.Equal(t, items[0], testNumberData[2])
		assert.Equal(t, items[1], testNumberData[1])
		assert.Equal(t, items[2], testNumberData[0])
	})

	t.Run("pk and sk are both bools", func(t *testing.T) {
		tableInfo := &models.TableInfo{Keys: models.KeyAttribute{PartitionKey: "pk", SortKey: "sk"}}

		items := make([]models.Item, len(testBoolData))
		copy(items, testBoolData)

		models.Sort(items, tableInfo)

		assert.Equal(t, items[0], testBoolData[2])
		assert.Equal(t, items[1], testBoolData[1])
		assert.Equal(t, items[2], testBoolData[0])
	})
}

var testStringData = []models.Item{
	{
		"pk":    &types.AttributeValueMemberS{Value: "bbb"},
		"sk":    &types.AttributeValueMemberS{Value: "131"},
		"beta":  &types.AttributeValueMemberN{Value: "2468"},
		"gamma": &types.AttributeValueMemberS{Value: "foobar"},
	},
	{
		"pk":    &types.AttributeValueMemberS{Value: "abc"},
		"sk":    &types.AttributeValueMemberS{Value: "111"},
		"alpha": &types.AttributeValueMemberS{Value: "This is some value"},
	},
	{
		"pk":    &types.AttributeValueMemberS{Value: "abc"},
		"sk":    &types.AttributeValueMemberS{Value: "222"},
		"alpha": &types.AttributeValueMemberS{Value: "This is another some value"},
		"beta":  &types.AttributeValueMemberN{Value: "2468"},
	},
}

var testNumberData = []models.Item{
	{
		"pk":    &types.AttributeValueMemberN{Value: "1141"},
		"sk":    &types.AttributeValueMemberN{Value: "1111"},
		"beta":  &types.AttributeValueMemberN{Value: "2468"},
		"gamma": &types.AttributeValueMemberS{Value: "foobar"},
	},
	{
		"pk":    &types.AttributeValueMemberN{Value: "1141"},
		"sk":    &types.AttributeValueMemberN{Value: "111.5"},
		"alpha": &types.AttributeValueMemberS{Value: "This is some value"},
	},
	{
		"pk":    &types.AttributeValueMemberN{Value: "5"},
		"sk":    &types.AttributeValueMemberN{Value: "222"},
		"alpha": &types.AttributeValueMemberS{Value: "This is another some value"},
		"beta":  &types.AttributeValueMemberN{Value: "2468"},
	},
}

var testBoolData = []models.Item{
	{
		"pk":    &types.AttributeValueMemberBOOL{Value: true},
		"sk":    &types.AttributeValueMemberBOOL{Value: true},
		"beta":  &types.AttributeValueMemberN{Value: "2468"},
		"gamma": &types.AttributeValueMemberS{Value: "foobar"},
	},
	{
		"pk":    &types.AttributeValueMemberBOOL{Value: true},
		"sk":    &types.AttributeValueMemberBOOL{Value: false},
		"alpha": &types.AttributeValueMemberS{Value: "This is some value"},
	},
	{
		"pk":    &types.AttributeValueMemberBOOL{Value: false},
		"sk":    &types.AttributeValueMemberBOOL{Value: false},
		"alpha": &types.AttributeValueMemberS{Value: "This is another some value"},
		"beta":  &types.AttributeValueMemberN{Value: "2468"},
	},
}
