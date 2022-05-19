package controllers_test

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/awstools/test/testdynamo"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
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

func TestTableReadController_ExportCSV(t *testing.T) {
	client, cleanupFn := testdynamo.SetupTestTable(t, testData)
	defer cleanupFn()

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider)
	readController := controllers.NewTableReadController(service, "alpha-table")

	t.Run("should export result set to CSV file", func(t *testing.T) {
		tempFile := tempFile(t)

		invokeCommand(t, readController.Init())
		invokeCommand(t, readController.ExportCSV(tempFile))

		bts, err := os.ReadFile(tempFile)
		assert.NoError(t, err)

		assert.Equal(t, string(bts), strings.Join([]string{
			"pk,sk,alpha,beta,gamma\n",
			"abc,111,This is some value,,\n",
			"abc,222,This is another some value,1231,\n",
			"bbb,131,,2468,foobar\n",
		}, ""))
	})

	t.Run("should return error if result set is not set", func(t *testing.T) {
		tempFile := tempFile(t)
		readController := controllers.NewTableReadController(service, "non-existant-table")

		invokeCommandExpectingError(t, readController.Init())
		invokeCommandExpectingError(t, readController.ExportCSV(tempFile))
	})

	// Hidden items?
}

func tempFile(t *testing.T) string {
	t.Helper()

	tempFile, err := os.CreateTemp("", "export.csv")
	assert.NoError(t, err)
	tempFile.Close()

	t.Cleanup(func() {
		os.Remove(tempFile.Name())
	})

	return tempFile.Name()
}

func invokeCommand(t *testing.T, cmd tea.Cmd) {
	msg := cmd()

	err, isErr := msg.(events.ErrorMsg)
	if isErr {
		assert.Fail(t, fmt.Sprintf("expected no error but got one: %v", err))
	}
}

func invokeCommandExpectingError(t *testing.T, cmd tea.Cmd) {
	msg := cmd()

	_, isErr := msg.(events.ErrorMsg)
	assert.True(t, isErr)
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
