package controllers_test

import (
	"testing"

	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/awstools/test/testdynamo"
)

func TestTableWriteController_ToggleReadWrite(t *testing.T) {
	t.Skip("needs to be updated")

	/*
		twc, _, closeFn := setupController(t)
		t.Cleanup(closeFn)

		t.Run("should enabling read write if disabled", func(t *testing.T) {
			ctx, uiCtx := testuictx.New(context.Background())
			ctx = controllers.ContextWithState(ctx, controllers.State{
				InReadWriteMode: false,
			})

			err := twc.ToggleReadWrite().Execute(ctx)
			assert.NoError(t, err)

			assert.Contains(t, uiCtx.Messages, controllers.SetReadWrite{NewValue: true})
		})

		t.Run("should disable read write if enabled", func(t *testing.T) {
			ctx, uiCtx := testuictx.New(context.Background())
			ctx = controllers.ContextWithState(ctx, controllers.State{
				InReadWriteMode: true,
			})

			err := twc.ToggleReadWrite().Execute(ctx)
			assert.NoError(t, err)

			assert.Contains(t, uiCtx.Messages, controllers.SetReadWrite{NewValue: false})
		})
	*/
}

func TestTableWriteController_Delete(t *testing.T) {
	/*
		t.Run("should delete selected item if in read/write mode is inactive", func(t *testing.T) {
			twc, ctrls, closeFn := setupController(t)
			t.Cleanup(closeFn)

			ti, err := ctrls.tableService.Describe(context.Background(), ctrls.tableName)
			assert.NoError(t, err)

			resultSet, err := ctrls.tableService.Scan(context.Background(), ti)
			assert.NoError(t, err)
			assert.Len(t, resultSet.Items, 3)

			ctx, uiCtx := testuictx.New(context.Background())
			ctx = controllers.ContextWithState(ctx, controllers.State{
				ResultSet:       resultSet,
				SelectedItem:    resultSet.Items[1],
				InReadWriteMode: true,
			})

			op := twc.Delete()

			// Should prompt first
			err = op.Execute(ctx)
			assert.NoError(t, err)

			_ = uiCtx

	*/
	/*
		promptRequest, ok := uiCtx.Messages[0].(events.PromptForInput)
		assert.True(t, ok)

		// After prompt, continue to delete
		err = promptRequest.OnDone.Execute(uimodels.WithPromptValue(ctx, "y"))
		assert.NoError(t, err)

		afterResultSet, err := ctrls.tableService.Scan(context.Background(), ti)
		assert.NoError(t, err)
		assert.Len(t, afterResultSet.Items, 2)
		assert.Contains(t, afterResultSet.Items, resultSet.Items[0])
		assert.NotContains(t, afterResultSet.Items, resultSet.Items[1])
		assert.Contains(t, afterResultSet.Items, resultSet.Items[2])
	*/
	/*
		})

		t.Run("should not delete selected item if prompt is not y", func(t *testing.T) {
			twc, ctrls, closeFn := setupController(t)
			t.Cleanup(closeFn)

			ti, err := ctrls.tableService.Describe(context.Background(), ctrls.tableName)
			assert.NoError(t, err)

			resultSet, err := ctrls.tableService.Scan(context.Background(), ti)
			assert.NoError(t, err)
			assert.Len(t, resultSet.Items, 3)

			ctx, uiCtx := testuictx.New(context.Background())
			ctx = controllers.ContextWithState(ctx, controllers.State{
				ResultSet:       resultSet,
				SelectedItem:    resultSet.Items[1],
				InReadWriteMode: true,
			})

			op := twc.Delete()

			// Should prompt first
			err = op.Execute(ctx)
			assert.NoError(t, err)
			_ = uiCtx
	*/
	/*
		promptRequest, ok := uiCtx.Messages[0].(events.PromptForInput)
		assert.True(t, ok)

		// After prompt, continue to delete
		err = promptRequest.OnDone.Execute(uimodels.WithPromptValue(ctx, "n"))
		assert.Error(t, err)

		afterResultSet, err := ctrls.tableService.Scan(context.Background(), ti)
		assert.NoError(t, err)
		assert.Len(t, afterResultSet.Items, 3)
		assert.Contains(t, afterResultSet.Items, resultSet.Items[0])
		assert.Contains(t, afterResultSet.Items, resultSet.Items[1])
		assert.Contains(t, afterResultSet.Items, resultSet.Items[2])
	*/
	/*
		})

		t.Run("should not delete if read/write mode is inactive", func(t *testing.T) {
			tableWriteController, ctrls, closeFn := setupController(t)
			t.Cleanup(closeFn)

			ti, err := ctrls.tableService.Describe(context.Background(), ctrls.tableName)
			assert.NoError(t, err)

			resultSet, err := ctrls.tableService.Scan(context.Background(), ti)
			assert.NoError(t, err)
			assert.Len(t, resultSet.Items, 3)

			ctx, _ := testuictx.New(context.Background())
			ctx = controllers.ContextWithState(ctx, controllers.State{
				ResultSet:       resultSet,
				SelectedItem:    resultSet.Items[1],
				InReadWriteMode: false,
			})

			op := tableWriteController.Delete()

			err = op.Execute(ctx)
			assert.Error(t, err)
		})

	*/
}

type controller struct {
	tableName    string
	tableService *tables.Service
}

func setupController(t *testing.T) (*controllers.TableWriteController, controller, func()) {
	tableName := "table-write-controller-table"

	client, cleanupFn := testdynamo.SetupTestTable(t, tableName, testData)
	provider := dynamo.NewProvider(client)
	tableService := tables.NewService(provider)
	tableReadController := controllers.NewTableReadController(tableService, tableName)
	tableWriteController := controllers.NewTableWriteController(tableService, tableReadController)
	return tableWriteController, controller{
		tableName:    tableName,
		tableService: tableService,
	}, cleanupFn
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
