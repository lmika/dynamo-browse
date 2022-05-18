package controllers_test

import (
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/awstools/test/testdynamo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTableReadController_InitTable(t *testing.T) {
	client, cleanupFn := testdynamo.SetupTestTable(t, testData)
	defer cleanupFn()

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider)

	t.Run("should prompt for table if no table name provided", func(t *testing.T) {
		readController := controllers.NewTableReadController(service, "")

		cmd := readController.Init()
		event := cmd()

		assert.IsType(t, controllers.PromptForTableMsg{}, event)
	})

	t.Run("should scan table if table name provided", func(t *testing.T) {
		readController := controllers.NewTableReadController(service, "")

		cmd := readController.Init()
		event := cmd()

		assert.IsType(t, controllers.PromptForTableMsg{}, event)
	})
}

func TestTableReadController_ListTables(t *testing.T) {
	client, cleanupFn := testdynamo.SetupTestTable(t, testData)
	defer cleanupFn()

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider)
	readController := controllers.NewTableReadController(service, "")

	t.Run("returns a list of tables", func(t *testing.T) {
		cmd := readController.ListTables()
		event := cmd().(controllers.PromptForTableMsg)

		assert.Equal(t, []string{"alpha-table", "bravo-table"}, event.Tables)

		selectedCmd := event.OnSelected("alpha-table")
		selectedEvent := selectedCmd()

		resultSet := selectedEvent.(controllers.NewResultSet)
		assert.Equal(t, "alpha-table", resultSet.ResultSet.TableInfo.Name)
		assert.Equal(t, "pk", resultSet.ResultSet.TableInfo.Keys.PartitionKey)
		assert.Equal(t, "sk", resultSet.ResultSet.TableInfo.Keys.SortKey)
	})
}

var testData = []testdynamo.TestData{
	{
		TableName: "alpha-table",
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
	{
		TableName: "bravo-table",
		Data: []map[string]interface{}{
			{
				"pk":    "foo",
				"sk":    "bar",
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
