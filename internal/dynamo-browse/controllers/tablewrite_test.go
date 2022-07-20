package controllers_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/awstools/test/testdynamo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTableWriteController_NewItem(t *testing.T) {
	t.Run("should add an item with pk and sk set at the end of the result set", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)

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

func TestTableWriteController_SetAttributeValue(t *testing.T) {
	t.Run("should preserve the type of the field if unspecified", func(t *testing.T) {

		scenarios := []struct {
			attrKey   string
			attrValue string
			expected  types.AttributeValue
		}{
			{
				attrKey:   "alpha",
				attrValue: "a new value",
				expected:  &types.AttributeValueMemberS{Value: "a new value"},
			},
			{
				attrKey:   "age",
				attrValue: "1234",
				expected:  &types.AttributeValueMemberN{Value: "1234"},
			},
			{
				attrKey:   "useMailing",
				attrValue: "t",
				expected:  &types.AttributeValueMemberBOOL{Value: true},
			},
			{
				attrKey:   "useMailing",
				attrValue: "f",
				expected:  &types.AttributeValueMemberBOOL{Value: false},
			},
		}

		for _, scenario := range scenarios {
			t.Run(fmt.Sprintf("should set value of field: %v", scenario.attrKey), func(t *testing.T) {
				client := testdynamo.SetupTestTable(t, testData)

				provider := dynamo.NewProvider(client)
				service := tables.NewService(provider)

				state := controllers.NewState()
				readController := controllers.NewTableReadController(state, service, "alpha-table")
				writeController := controllers.NewTableWriteController(state, service, readController)

				invokeCommand(t, readController.Init())
				invokeCommandWithPrompt(t, writeController.SetAttributeValue(0, models.UnsetItemType, scenario.attrKey), scenario.attrValue)

				after, _ := state.ResultSet().Items()[0][scenario.attrKey]
				assert.Equal(t, scenario.expected, after)
				assert.True(t, state.ResultSet().IsDirty(0))
			})
		}
	})

	t.Run("should use type of selected item for marked fields if unspecified", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		invokeCommand(t, readController.Init())
		invokeCommand(t, writeController.ToggleMark(0))
		invokeCommandWithPrompt(t, writeController.SetAttributeValue(1, models.UnsetItemType, "alpha"), "a brand new value")

		after1 := state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value
		assert.Equal(t, "a brand new value", after1)
		assert.True(t, state.ResultSet().IsDirty(0))
		assert.False(t, state.ResultSet().IsDirty(1))
	})

	t.Run("should change the value to a particular field if already present", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		scenarios := []struct {
			attrType  models.ItemType
			attrValue string
			expected  types.AttributeValue
		}{
			{
				attrType:  models.StringItemType,
				attrValue: "a new value",
				expected:  &types.AttributeValueMemberS{Value: "a new value"},
			},
			{
				attrType:  models.NumberItemType,
				attrValue: "1234",
				expected:  &types.AttributeValueMemberN{Value: "1234"},
			},
			{
				attrType:  models.BoolItemType,
				attrValue: "true",
				expected:  &types.AttributeValueMemberBOOL{Value: true},
			},
			{
				attrType:  models.BoolItemType,
				attrValue: "false",
				expected:  &types.AttributeValueMemberBOOL{Value: false},
			},
			{
				attrType:  models.NullItemType,
				attrValue: "",
				expected:  &types.AttributeValueMemberNULL{Value: true},
			},
		}

		for _, scenario := range scenarios {
			t.Run(fmt.Sprintf("should change the value of a field to type %v", scenario.attrType), func(t *testing.T) {
				state := controllers.NewState()
				readController := controllers.NewTableReadController(state, service, "alpha-table")
				writeController := controllers.NewTableWriteController(state, service, readController)

				invokeCommand(t, readController.Init())
				before, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
				assert.Equal(t, "This is some value", before)
				assert.False(t, state.ResultSet().IsDirty(0))

				if scenario.attrValue == "" {
					invokeCommand(t, writeController.SetAttributeValue(0, scenario.attrType, "alpha"))
				} else {
					invokeCommandWithPrompt(t, writeController.SetAttributeValue(0, scenario.attrType, "alpha"), scenario.attrValue)
				}

				after, _ := state.ResultSet().Items()[0]["alpha"]
				assert.Equal(t, scenario.expected, after)
				assert.True(t, state.ResultSet().IsDirty(0))
			})

			t.Run(fmt.Sprintf("should change value of nested field to type %v", scenario.attrType), func(t *testing.T) {
				state := controllers.NewState()
				readController := controllers.NewTableReadController(state, service, "alpha-table")
				writeController := controllers.NewTableWriteController(state, service, readController)

				invokeCommand(t, readController.Init())

				beforeAddress := state.ResultSet().Items()[0]["address"].(*types.AttributeValueMemberM)
				beforeStreet := beforeAddress.Value["street"].(*types.AttributeValueMemberS).Value

				assert.Equal(t, "Fake st.", beforeStreet)
				assert.False(t, state.ResultSet().IsDirty(0))

				if scenario.attrValue == "" {
					invokeCommand(t, writeController.SetAttributeValue(0, scenario.attrType, "address.street"))
				} else {
					invokeCommandWithPrompt(t, writeController.SetAttributeValue(0, scenario.attrType, "address.street"), scenario.attrValue)
				}

				afterAddress := state.ResultSet().Items()[0]["address"].(*types.AttributeValueMemberM)
				after := afterAddress.Value["street"]

				assert.Equal(t, scenario.expected, after)
				assert.True(t, state.ResultSet().IsDirty(0))
			})
		}
	})
}

func TestTableWriteController_DeleteAttribute(t *testing.T) {
	client := testdynamo.SetupTestTable(t, testData)

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
		client := testdynamo.SetupTestTable(t, testData)

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
		invokeCommandWithPrompt(t, writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandWithPrompt(t, writeController.PutItem(0), "y")

		// Rescan the table
		invokeCommand(t, readController.Rescan())
		after, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "a new value", after)
		assert.False(t, state.ResultSet().IsDirty(0))
	})

	t.Run("should not put the selected item if user does not confirm", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)

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
		invokeCommandWithPrompt(t, writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandWithPrompt(t, writeController.PutItem(0), "n")

		current, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "a new value", current)
		assert.True(t, state.ResultSet().IsDirty(0))

		// Rescan the table to confirm item is not modified
		invokeCommandWithPrompt(t, readController.Rescan(), "y")
		after, _ := state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", after)
		assert.False(t, state.ResultSet().IsDirty(0))
	})

	t.Run("should not put the selected item if not dirty", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)

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

func TestTableWriteController_PutItems(t *testing.T) {
	t.Run("should put all dirty items if none are marked", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		invokeCommand(t, readController.Init())

		// Modify the item and put it
		invokeCommandWithPrompt(t, writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandWithPrompt(t, writeController.SetAttributeValue(2, models.StringItemType, "alpha"), "another new value")

		invokeCommandWithPrompt(t, writeController.PutItems(), "y")

		// Rescan the table
		invokeCommand(t, readController.Rescan())

		assert.Equal(t, "a new value", state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "another new value", state.ResultSet().Items()[2]["alpha"].(*types.AttributeValueMemberS).Value)

		assert.False(t, state.ResultSet().IsDirty(0))
		assert.False(t, state.ResultSet().IsDirty(2))
	})

	t.Run("only put marked items", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		invokeCommand(t, readController.Init())

		// Modify the item and put it
		invokeCommandWithPrompt(t, writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandWithPrompt(t, writeController.SetAttributeValue(2, models.StringItemType, "alpha"), "another new value")
		invokeCommand(t, writeController.ToggleMark(0))

		invokeCommandWithPrompt(t, writeController.PutItems(), "y")

		// Verify dirty items are unchanged
		assert.Equal(t, "a new value", state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "another new value", state.ResultSet().Items()[2]["alpha"].(*types.AttributeValueMemberS).Value)

		assert.False(t, state.ResultSet().IsDirty(0))
		assert.True(t, state.ResultSet().IsDirty(2))

		// Rescan the table and verify dirty items were not written
		invokeCommandWithPrompt(t, readController.Rescan(), "y")

		assert.Equal(t, "a new value", state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Nil(t, state.ResultSet().Items()[2]["alpha"])

		assert.False(t, state.ResultSet().IsDirty(0))
		assert.False(t, state.ResultSet().IsDirty(2))
	})

	t.Run("do not put marked items which are not diry", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)

		provider := dynamo.NewProvider(client)
		service := tables.NewService(provider)

		state := controllers.NewState()
		readController := controllers.NewTableReadController(state, service, "alpha-table")
		writeController := controllers.NewTableWriteController(state, service, readController)

		invokeCommand(t, readController.Init())

		// Modify the item and put it
		invokeCommandWithPrompt(t, writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandWithPrompt(t, writeController.SetAttributeValue(2, models.StringItemType, "alpha"), "another new value")
		invokeCommand(t, writeController.ToggleMark(1))

		invokeCommand(t, writeController.PutItems())

		// Verify dirty items are unchanged
		assert.Equal(t, "a new value", state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "another new value", state.ResultSet().Items()[2]["alpha"].(*types.AttributeValueMemberS).Value)

		assert.True(t, state.ResultSet().IsDirty(0))
		assert.True(t, state.ResultSet().IsDirty(2))

		// Rescan the table and verify dirty items were not written
		invokeCommandWithPrompt(t, readController.Rescan(), "y")

		assert.Equal(t, "This is some value", state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Nil(t, state.ResultSet().Items()[2]["alpha"])

		assert.False(t, state.ResultSet().IsDirty(0))
		assert.False(t, state.ResultSet().IsDirty(2))
	})
}

func TestTableWriteController_TouchItem(t *testing.T) {
	t.Run("should put the selected item if unmodified", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)

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
		client := testdynamo.SetupTestTable(t, testData)

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
		invokeCommandWithPrompt(t, writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandExpectingError(t, writeController.TouchItem(0))
	})
}

func TestTableWriteController_NoisyTouchItem(t *testing.T) {
	t.Run("should delete and put the selected item if unmodified", func(t *testing.T) {
		client := testdynamo.SetupTestTable(t, testData)

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
		client := testdynamo.SetupTestTable(t, testData)

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
		invokeCommandWithPrompt(t, writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandExpectingError(t, writeController.NoisyTouchItem(0))
	})
}
