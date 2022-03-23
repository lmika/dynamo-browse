package tables_test

import (
	"context"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/awstools/test/testdynamo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_Scan(t *testing.T) {
	tableName := "service-scan-test-table"

	client, cleanupFn := testdynamo.SetupTestTable(t, tableName, testData)
	defer cleanupFn()
	provider := dynamo.NewProvider(client)

	t.Run("return all columns and fields in sorted order", func(t *testing.T) {
		ctx := context.Background()

		service := tables.NewService(provider)
		rs, err := service.Scan(ctx, tableName)
		assert.NoError(t, err)

		// Hash first, then range, then columns in alphabetic order
		assert.Equal(t, rs.Table, tableName)
		assert.Equal(t, rs.Columns, []string{"pk", "sk", "alpha", "beta", "gamma"})
		assert.Equal(t, rs.Items[0], testdynamo.TestRecordAsItem(t, testData[1]))
		assert.Equal(t, rs.Items[1], testdynamo.TestRecordAsItem(t, testData[0]))
		assert.Equal(t, rs.Items[2], testdynamo.TestRecordAsItem(t, testData[2]))
	})
}

var testData = testdynamo.TestData{
	{
		"pk":    "abc",
		"sk":    "222",
		"alpha": "This is another some value",
		"beta":  1231,
	},
	{
		"pk":    "abc",
		"sk":    "111",
		"alpha": "This is some value",
	},
	{
		"pk":    "bbb",
		"sk":    "131",
		"beta":  2468,
		"gamma": "foobar",
	},
}
