package controllers_test

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/awstools/test/testdynamo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTableWriteController_NewItem(t *testing.T) {
	t.Run("should add an item with pk and sk set at the end of the result set", func(t *testing.T) {
		client, cleanupFn := testdynamo.SetupTestTable(t, testData)
		defer cleanupFn()

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		invokeCommand(t, readController.Init())
		assert.Len(t, state.ResultSet().Items(), 3)

		// Prompt for keys
		invokeCommandWithPrompts(t, writeController.NewItem(), "pk-value", "sk-value")

		newResultSet := state.ResultSet()
		assert.Len(t, newResultSet.Items(), 4)
		assert.Len(t, newResultSet.Items()[3], 2)

		pk, _ := newResultSet.Items()[3].AttributeValueAsString("pk")
		sk, _ := newResultSet.Items()[3].AttributeValueAsString("sk")
		assert.Equal(t, "pk-value", pk)
		assert.Equal(t, "sk-value", sk)
		assert.True(t, newResultSet.IsNew(3))
		assert.True(t, newResultSet.IsDirty(3))
	})
}

func TestTableWriteController_SetStringValue(t *testing.T) {
	client, cleanupFn := testdynamo.SetupTestTable(t, testData)
	defer cleanupFn()

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider)

	t.Run("should change the value of a string field if already present", func(t *testing.T) {
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

	t.Run("should change the value of a string field within a map if already present", func(t *testing.T) {
		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		invokeCommand(t, readController.Init())

		beforeAddress := state.ResultSet().Items()[0]["address"].(*types.AttributeValueMemberM)
		beforeStreet := beforeAddress.Value["street"].(*types.AttributeValueMemberS).Value

		assert.Equal(t, "Fake st.", beforeStreet)
		assert.False(t, state.ResultSet().IsDirty(0))

		invokeCommandWithPrompt(t, writeController.SetStringValue(0, "address.street"), "Fiction rd.")

		afterAddress := state.ResultSet().Items()[0]["address"].(*types.AttributeValueMemberM)
		afterStreet := afterAddress.Value["street"].(*types.AttributeValueMemberS).Value

		assert.Equal(t, "Fiction rd.", afterStreet)
		assert.True(t, state.ResultSet().IsDirty(0))
	})
}

func TestTableWriteController_SetNumberValue(t *testing.T) {
	client, cleanupFn := testdynamo.SetupTestTable(t, testData)
	defer cleanupFn()

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider)

	t.Run("should change the value of a number field if already present", func(t *testing.T) {
		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		invokeCommand(t, readController.Init())
		before, _ := state.ResultSet().Items()[0].AttributeValueAsString("age")
		assert.Equal(t, "23", before)
		assert.False(t, state.ResultSet().IsDirty(0))

		invokeCommandWithPrompt(t, writeController.SetNumberValue(0, "age"), "46")

		after, _ := state.ResultSet().Items()[0].AttributeValueAsString("age")
		assert.Equal(t, "46", after)
		assert.True(t, state.ResultSet().IsDirty(0))
	})

	t.Run("should change the value of a number field within a map if already present", func(t *testing.T) {
		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		invokeCommand(t, readController.Init())

		beforeAddress := state.ResultSet().Items()[0]["address"].(*types.AttributeValueMemberM)
		beforeStreet := beforeAddress.Value["no"].(*types.AttributeValueMemberN).Value

		assert.Equal(t, "123", beforeStreet)
		assert.False(t, state.ResultSet().IsDirty(0))

		invokeCommandWithPrompt(t, writeController.SetNumberValue(0, "address.no"), "456")

		afterAddress := state.ResultSet().Items()[0]["address"].(*types.AttributeValueMemberM)
		afterStreet := afterAddress.Value["no"].(*types.AttributeValueMemberN).Value

		assert.Equal(t, "456", afterStreet)
		assert.True(t, state.ResultSet().IsDirty(0))
	})
}

func TestTableWriteController_DeleteAttribute(t *testing.T) {
	client, cleanupFn := testdynamo.SetupTestTable(t, testData)
	defer cleanupFn()

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider)

	t.Run("should delete top level attribute", func(t *testing.T) {
		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		invokeCommand(t, readController.Init())
		before, _ := state.ResultSet().Items()[0].AttributeValueAsString("age")
		assert.Equal(t, "23", before)
		assert.False(t, state.ResultSet().IsDirty(0))

		invokeCommand(t, writeController.DeleteAttribute(0, "age"))

		_, hasAge := state.ResultSet().Items()[0]["age"]
		assert.False(t, hasAge)
	})

	t.Run("should delete attribute of map", func(t *testing.T) {
		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		invokeCommand(t, readController.Init())

		beforeAddress := state.ResultSet().Items()[0]["address"].(*types.AttributeValueMemberM)
		beforeStreet := beforeAddress.Value["no"].(*types.AttributeValueMemberN).Value

		assert.Equal(t, "123", beforeStreet)
		assert.False(t, state.ResultSet().IsDirty(0))

		invokeCommand(t, writeController.DeleteAttribute(0, "address.no"))

		afterAddress := state.ResultSet().Items()[0]["address"].(*types.AttributeValueMemberM)
		_, hasStreet := afterAddress.Value["no"]

		assert.False(t, hasStreet)
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
