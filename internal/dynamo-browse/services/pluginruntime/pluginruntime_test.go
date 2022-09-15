package pluginruntime_test

import (
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/lmika/audax/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/audax/internal/dynamo-browse/providers/workspacestore"
	"github.com/lmika/audax/internal/dynamo-browse/services/pluginruntime"
	"github.com/lmika/audax/internal/dynamo-browse/services/tables"
	workspaces_service "github.com/lmika/audax/internal/dynamo-browse/services/workspaces"
	"github.com/lmika/audax/test/testdynamo"
	"github.com/lmika/audax/test/testworkspace"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPluginRuntime_MissingCommand(t *testing.T) {
	srv := setupTestService(t)

	pluginChan, err := srv.LoadScript("test.js", `
		console.log("Hello, world");
	`)
	assert.NoError(t, err)

	plugin, err := (<-pluginChan).Get()
	assert.NoError(t, err)
	_ = plugin
}

type testService struct {
	*pluginruntime.Service
}

func setupTestService(t *testing.T) *testService {
	client := testdynamo.SetupTestTable(t, testData)
	resultSetSnapshotStore := workspacestore.NewResultSetSnapshotStore(testworkspace.New(t))
	workspaceService := workspaces_service.NewService(resultSetSnapshotStore)

	state := controllers.NewState()
	service := tables.NewService(dynamo.NewProvider(client))

	srv := pluginruntime.New(state, service, workspaceService)
	srv.Start()

	return &testService{srv}
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
