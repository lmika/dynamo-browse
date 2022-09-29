package controllers_test

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/lmika/audax/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/audax/internal/dynamo-browse/providers/workspacestore"
	"github.com/lmika/audax/internal/dynamo-browse/services/itemrenderer"
	"github.com/lmika/audax/internal/dynamo-browse/services/tables"
	workspaces_service "github.com/lmika/audax/internal/dynamo-browse/services/workspaces"
	"github.com/lmika/audax/test/testdynamo"
	"github.com/lmika/audax/test/testworkspace"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestTableReadController_InitTable(t *testing.T) {
	client := testdynamo.SetupTestTable(t, testData)

	resultSetSnapshotStore := workspacestore.NewResultSetSnapshotStore(testworkspace.New(t))
	workspaceService := workspaces_service.NewService(resultSetSnapshotStore)
	itemRendererService := itemrenderer.NewService(itemrenderer.PlainTextRenderer(), itemrenderer.PlainTextRenderer())

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider, &mockedSetting{})

	t.Run("should prompt for table if no table name provided", func(t *testing.T) {
		readController := controllers.NewTableReadController(controllers.NewState(), service, workspaceService, itemRendererService, "", false)

		event := readController.Init()

		assert.IsType(t, controllers.PromptForTableMsg{}, event)
	})

	t.Run("should scan table if table name provided", func(t *testing.T) {
		readController := controllers.NewTableReadController(controllers.NewState(), service, workspaceService, itemRendererService, "", false)

		event := readController.Init()

		assert.IsType(t, controllers.PromptForTableMsg{}, event)
	})
}

func TestTableReadController_ListTables(t *testing.T) {
	client := testdynamo.SetupTestTable(t, testData)

	resultSetSnapshotStore := workspacestore.NewResultSetSnapshotStore(testworkspace.New(t))
	workspaceService := workspaces_service.NewService(resultSetSnapshotStore)
	itemRendererService := itemrenderer.NewService(itemrenderer.PlainTextRenderer(), itemrenderer.PlainTextRenderer())

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider, &mockedSetting{})
	readController := controllers.NewTableReadController(controllers.NewState(), service, workspaceService, itemRendererService, "", false)

	t.Run("returns a list of tables", func(t *testing.T) {
		event := readController.ListTables().(controllers.PromptForTableMsg)

		assert.Equal(t, []string{"alpha-table", "bravo-table"}, event.Tables)

		selectedEvent := event.OnSelected("alpha-table")

		resultSet := selectedEvent.(controllers.NewResultSet)
		assert.Equal(t, "alpha-table", resultSet.ResultSet.TableInfo.Name)
		assert.Equal(t, "pk", resultSet.ResultSet.TableInfo.Keys.PartitionKey)
		assert.Equal(t, "sk", resultSet.ResultSet.TableInfo.Keys.SortKey)
	})
}

func TestTableReadController_Rescan(t *testing.T) {
	client := testdynamo.SetupTestTable(t, testData)

	resultSetSnapshotStore := workspacestore.NewResultSetSnapshotStore(testworkspace.New(t))
	workspaceService := workspaces_service.NewService(resultSetSnapshotStore)
	itemRendererService := itemrenderer.NewService(itemrenderer.PlainTextRenderer(), itemrenderer.PlainTextRenderer())

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider, &mockedSetting{})
	state := controllers.NewState()
	readController := controllers.NewTableReadController(state, service, workspaceService, itemRendererService, "bravo-table", false)

	t.Run("should perform a rescan", func(t *testing.T) {
		invokeCommand(t, readController.Init())
		invokeCommand(t, readController.Rescan())
	})

	t.Run("should prompt to rescan if any dirty rows", func(t *testing.T) {
		invokeCommand(t, readController.Init())

		state.ResultSet().SetDirty(0, true)

		invokeCommandWithPrompt(t, readController.Rescan(), "y")

		assert.False(t, state.ResultSet().IsDirty(0))
	})

	t.Run("should not rescan if any dirty rows", func(t *testing.T) {
		invokeCommand(t, readController.Init())

		state.ResultSet().SetDirty(0, true)

		invokeCommandWithPrompt(t, readController.Rescan(), "n")

		assert.True(t, state.ResultSet().IsDirty(0))
	})
}

func TestTableReadController_ExportCSV(t *testing.T) {
	client := testdynamo.SetupTestTable(t, testData)

	resultSetSnapshotStore := workspacestore.NewResultSetSnapshotStore(testworkspace.New(t))
	workspaceService := workspaces_service.NewService(resultSetSnapshotStore)
	itemRendererService := itemrenderer.NewService(itemrenderer.PlainTextRenderer(), itemrenderer.PlainTextRenderer())

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider, &mockedSetting{})
	readController := controllers.NewTableReadController(controllers.NewState(), service, workspaceService, itemRendererService, "bravo-table", false)

	t.Run("should export result set to CSV file", func(t *testing.T) {
		tempFile := tempFile(t)

		invokeCommand(t, readController.Init())
		invokeCommand(t, readController.ExportCSV(tempFile))

		bts, err := os.ReadFile(tempFile)
		assert.NoError(t, err)

		assert.Equal(t, string(bts), strings.Join([]string{
			"pk,sk,alpha,beta,gamma\n",
			"abc,222,This is another some value,1231,\n",
			"bbb,131,,2468,foobar\n",
			"foo,bar,This is some value,,\n",
		}, ""))
	})

	t.Run("should return error if result set is not set", func(t *testing.T) {
		tempFile := tempFile(t)
		readController := controllers.NewTableReadController(controllers.NewState(), service, workspaceService, itemRendererService, "non-existant-table", false)

		invokeCommandExpectingError(t, readController.Init())
		invokeCommandExpectingError(t, readController.ExportCSV(tempFile))
	})

	// Hidden items?
}

func TestTableReadController_Query(t *testing.T) {
	client := testdynamo.SetupTestTable(t, testData)

	resultSetSnapshotStore := workspacestore.NewResultSetSnapshotStore(testworkspace.New(t))
	workspaceService := workspaces_service.NewService(resultSetSnapshotStore)
	itemRendererService := itemrenderer.NewService(itemrenderer.PlainTextRenderer(), itemrenderer.PlainTextRenderer())

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider, &mockedSetting{})
	readController := controllers.NewTableReadController(controllers.NewState(), service, workspaceService, itemRendererService, "bravo-table", false)

	t.Run("should run scan with filter based on user query", func(t *testing.T) {
		tempFile := tempFile(t)

		invokeCommand(t, readController.Init())
		invokeCommandWithPrompts(t, readController.PromptForQuery(), `pk ^= "abc"`)
		invokeCommand(t, readController.ExportCSV(tempFile))

		bts, err := os.ReadFile(tempFile)
		assert.NoError(t, err)

		assert.Equal(t, string(bts), strings.Join([]string{
			"pk,sk,alpha,beta\n",
			"abc,222,This is another some value,1231\n",
		}, ""))
	})

	t.Run("should return error if result set is not set", func(t *testing.T) {
		tempFile := tempFile(t)
		readController := controllers.NewTableReadController(controllers.NewState(), service, workspaceService, itemRendererService, "non-existant-table", false)

		invokeCommandExpectingError(t, readController.Init())
		invokeCommandExpectingError(t, readController.ExportCSV(tempFile))
	})
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

//
//func testWorkspace(t *testing.T) *workspaces.Workspace {
//	wsTempFile := tempFile(t)
//
//	wsManager := workspaces.New(workspaces.MetaInfo{Command: "dynamo-browse"})
//	ws, err := wsManager.Open(wsTempFile)
//	if err != nil {
//		t.Fatalf("cannot create workspace manager: %v", err)
//	}
//	t.Cleanup(func() { ws.Close() })
//
//	return ws
//}

func invokeCommand(t *testing.T, msg tea.Msg) tea.Msg {
	err, isErr := msg.(events.ErrorMsg)
	if isErr {
		assert.Fail(t, fmt.Sprintf("expected no error but got one: %v", err))
	}
	return msg
}

func invokeCommandWithPrompt(t *testing.T, msg tea.Msg, promptValue string) {
	pi, isPi := msg.(events.PromptForInputMsg)
	if !isPi {
		assert.Fail(t, fmt.Sprintf("expected prompt for input but didn't get one"))
	}

	invokeCommand(t, pi.OnDone(promptValue))
}

func invokeCommandWithPrompts(t *testing.T, msg tea.Msg, promptValues ...string) {
	for _, promptValue := range promptValues {
		pi, isPi := msg.(events.PromptForInputMsg)
		if !isPi {
			assert.Fail(t, fmt.Sprintf("expected prompt for input but didn't get one: %T", msg))
		}

		msg = invokeCommand(t, pi.OnDone(promptValue))
	}
}

func invokeCommandWithPromptsExpectingError(t *testing.T, msg tea.Msg, promptValues ...string) {
	for _, promptValue := range promptValues {
		pi, isPi := msg.(events.PromptForInputMsg)
		if !isPi {
			assert.Fail(t, fmt.Sprintf("expected prompt for input but didn't get one"))
		}

		msg = invokeCommand(t, pi.OnDone(promptValue))
	}

	_, isErr := msg.(events.ErrorMsg)
	assert.True(t, isErr)
}

func invokeCommandExpectingError(t *testing.T, msg tea.Msg) {
	_, isErr := msg.(events.ErrorMsg)
	assert.True(t, isErr)
}

var testData = []testdynamo.TestData{
	{
		TableName: "alpha-table",
		Data: []map[string]interface{}{
			{
				"pk":         "abc",
				"sk":         "111",
				"alpha":      "This is some value",
				"age":        23,
				"useMailing": true,
				"address": map[string]any{
					"no":     123,
					"street": "Fake st.",
				},
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
