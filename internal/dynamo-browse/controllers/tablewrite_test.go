package controllers_test

import (
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/awstools/test/testdynamo"
	"github.com/stretchr/testify/assert"
	"testing"
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

/*
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
*/

func TestTableWriteController_NewItem(t *testing.T) {
	client, cleanupFn := testdynamo.SetupTestTable(t, testData)
	defer cleanupFn()

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider)

	t.Run("should add a new empty item at the end of the result set", func(t *testing.T) {
		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		invokeCommand(t, readController.Init())
		assert.Len(t, state.ResultSet().Items(), 3)

		invokeCommand(t, writeController.NewItem())
		newResultSet := state.ResultSet()
		assert.Len(t, newResultSet.Items(), 4)
		assert.Len(t, newResultSet.Items()[3], 0)
		assert.True(t, newResultSet.IsNew(3))
		assert.True(t, newResultSet.IsDirty(3))
	})
}

func TestTableWriteController_SetStringValue(t *testing.T) {
	client, cleanupFn := testdynamo.SetupTestTable(t, testData)
	defer cleanupFn()

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider)

	t.Run("should add a new empty item at the end of the result set", func(t *testing.T) {
		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		invokeCommand(t, readController.Init())
		before, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, state.ResultSet().IsDirty(0))

		invokeCommandWithPrompt(t, writeController.SetStringValue(0, "alpha"), "a new value")

		after, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "a new value", after)
		assert.True(t, state.ResultSet().IsDirty(0))
	})

	t.Run("should prevent duplicate partition,sort keys", func(t *testing.T) {
		t.Skip("TODO")
	})
}

func TestTableWriteController_PutItem(t *testing.T) {
	t.Run("should put the selected item if dirty", func(t *testing.T) {
		client, cleanupFn := testdynamo.SetupTestTable(t, testData)
		defer cleanupFn()

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		// Read the table
		invokeCommand(t, readController.Init())
		before, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, state.ResultSet().IsDirty(0))

		// Modify the item and put it
		invokeCommandWithPrompt(t, writeController.SetStringValue(0, "alpha"), "a new value")
		invokeCommandWithPrompt(t, writeController.PutItem(0), "y")

		// Rescan the table
		invokeCommand(t, readController.Rescan())
		after, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "a new value", after)
		assert.False(t, state.ResultSet().IsDirty(0))
	})

	t.Run("should not put the selected item if user does not confirm", func(t *testing.T) {
		client, cleanupFn := testdynamo.SetupTestTable(t, testData)
		defer cleanupFn()

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		// Read the table
		invokeCommand(t, readController.Init())
		before, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, state.ResultSet().IsDirty(0))

		// Modify the item but do not put it
		invokeCommandWithPrompt(t, writeController.SetStringValue(0, "alpha"), "a new value")
		invokeCommandWithPrompt(t, writeController.PutItem(0), "n")

		current, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "a new value", current)
		assert.True(t, state.ResultSet().IsDirty(0))

		// Rescan the table to confirm item is not modified
		invokeCommand(t, readController.Rescan())
		after, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", after)
		assert.False(t, state.ResultSet().IsDirty(0))
	})

	t.Run("should not put the selected item if not dirty", func(t *testing.T) {
		client, cleanupFn := testdynamo.SetupTestTable(t, testData)
		defer cleanupFn()

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		// Read the table
		invokeCommand(t, readController.Init())
		before, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, state.ResultSet().IsDirty(0))

		invokeCommandExpectingError(t, writeController.PutItem(0))
	})
}

func TestTableWriteController_TouchItem(t *testing.T) {
	t.Run("should put the selected item if unmodified", func(t *testing.T) {
		client, cleanupFn := testdynamo.SetupTestTable(t, testData)
		defer cleanupFn()

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		// Read the table
		invokeCommand(t, readController.Init())
		before, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, state.ResultSet().IsDirty(0))

		// Modify the item and put it
		invokeCommandWithPrompt(t, writeController.TouchItem(0), "y")

		// Rescan the table
		invokeCommand(t, readController.Rescan())
		after, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", after)
		assert.False(t, state.ResultSet().IsDirty(0))
	})

	t.Run("should not put the selected item if modified", func(t *testing.T) {
		client, cleanupFn := testdynamo.SetupTestTable(t, testData)
		defer cleanupFn()

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		// Read the table
		invokeCommand(t, readController.Init())
		before, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, state.ResultSet().IsDirty(0))

		// Modify the item and put it
		invokeCommandWithPrompt(t, writeController.SetStringValue(0, "alpha"), "a new value")
		invokeCommandExpectingError(t, writeController.TouchItem(0))
	})
}

func TestTableWriteController_NoisyTouchItem(t *testing.T) {
	t.Run("should delete and put the selected item if unmodified", func(t *testing.T) {
		client, cleanupFn := testdynamo.SetupTestTable(t, testData)
		defer cleanupFn()

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		// Read the table
		invokeCommand(t, readController.Init())
		before, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, state.ResultSet().IsDirty(0))

		// Modify the item and put it
		invokeCommandWithPrompt(t, writeController.NoisyTouchItem(0), "y")

		// Rescan the table
		invokeCommand(t, readController.Rescan())
		after, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", after)
		assert.False(t, state.ResultSet().IsDirty(0))
	})

	t.Run("should not put the selected item if modified", func(t *testing.T) {
		client, cleanupFn := testdynamo.SetupTestTable(t, testData)
		defer cleanupFn()

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		// Read the table
		invokeCommand(t, readController.Init())
		before, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, state.ResultSet().IsDirty(0))

		// Modify the item and put it
		invokeCommandWithPrompt(t, writeController.SetStringValue(0, "alpha"), "a new value")
		invokeCommandExpectingError(t, writeController.NoisyTouchItem(0))
	})
}
