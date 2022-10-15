package controllers_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/audax/internal/dynamo-browse/providers/settingstore"
	"github.com/lmika/audax/internal/dynamo-browse/providers/workspacestore"
	"github.com/lmika/audax/internal/dynamo-browse/services/itemrenderer"
	"github.com/lmika/audax/internal/dynamo-browse/services/jobs"
	"github.com/lmika/audax/internal/dynamo-browse/services/tables"
	workspaces_service "github.com/lmika/audax/internal/dynamo-browse/services/workspaces"
	"github.com/lmika/audax/test/testdynamo"
	"github.com/lmika/audax/test/testworkspace"
	bus "github.com/lmika/events"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTableWriteController_NewItem(t *testing.T) {
	t.Run("should add an item with pk and sk set at the end of the result set", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		invokeCommand(t, srv.readController.Init())
		assert.Len(t, srv.state.ResultSet().Items(), 3)

		// Prompt for keys
		invokeCommandWithPrompts(t, srv.writeController.NewItem(), "pk-value", "sk-value")

		newResultSet := srv.state.ResultSet()
		assert.Len(t, newResultSet.Items(), 4)
		assert.Len(t, newResultSet.Items()[3], 2)

		pk, _ := newResultSet.Items()[3].AttributeValueAsString("pk")
		sk, _ := newResultSet.Items()[3].AttributeValueAsString("sk")
		assert.Equal(t, "pk-value", pk)
		assert.Equal(t, "sk-value", sk)
		assert.True(t, newResultSet.IsNew(3))
		assert.True(t, newResultSet.IsDirty(3))
	})

	t.Run("should do nothing when in read-only mode", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table", isReadOnly: true})

		invokeCommand(t, srv.readController.Init())
		assert.Len(t, srv.state.ResultSet().Items(), 3)

		// Prompt for keys
		invokeCommandExpectingError(t, srv.writeController.NewItem())

		// ConfirmYes no changes
		invokeCommand(t, srv.readController.Rescan())

		newResultSet := srv.state.ResultSet()
		assert.Len(t, newResultSet.Items(), 3)
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
				srv := newService(t, serviceConfig{tableName: "alpha-table"})

				invokeCommand(t, srv.readController.Init())
				invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(0, models.UnsetItemType, scenario.attrKey), scenario.attrValue)

				after, _ := srv.state.ResultSet().Items()[0][scenario.attrKey]
				assert.Equal(t, scenario.expected, after)
				assert.True(t, srv.state.ResultSet().IsDirty(0))
			})
		}
	})

	t.Run("should use type of selected item for marked fields if unspecified", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		invokeCommand(t, srv.readController.Init())
		invokeCommand(t, srv.writeController.ToggleMark(0))
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(1, models.UnsetItemType, "alpha"), "a brand new value")

		after1 := srv.state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value
		assert.Equal(t, "a brand new value", after1)
		assert.True(t, srv.state.ResultSet().IsDirty(0))
		assert.False(t, srv.state.ResultSet().IsDirty(1))
	})

	t.Run("should change the value to a particular field if already present", func(t *testing.T) {
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
				srv := newService(t, serviceConfig{tableName: "alpha-table"})

				invokeCommand(t, srv.readController.Init())
				before, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
				assert.Equal(t, "This is some value", before)
				assert.False(t, srv.state.ResultSet().IsDirty(0))

				if scenario.attrValue == "" {
					invokeCommand(t, srv.writeController.SetAttributeValue(0, scenario.attrType, "alpha"))
				} else {
					invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(0, scenario.attrType, "alpha"), scenario.attrValue)
				}

				after, _ := srv.state.ResultSet().Items()[0]["alpha"]
				assert.Equal(t, scenario.expected, after)
				assert.True(t, srv.state.ResultSet().IsDirty(0))
			})

			t.Run(fmt.Sprintf("should change value of nested field to type %v", scenario.attrType), func(t *testing.T) {
				srv := newService(t, serviceConfig{tableName: "alpha-table"})

				invokeCommand(t, srv.readController.Init())

				beforeAddress := srv.state.ResultSet().Items()[0]["address"].(*types.AttributeValueMemberM)
				beforeStreet := beforeAddress.Value["street"].(*types.AttributeValueMemberS).Value

				assert.Equal(t, "Fake st.", beforeStreet)
				assert.False(t, srv.state.ResultSet().IsDirty(0))

				if scenario.attrValue == "" {
					invokeCommand(t, srv.writeController.SetAttributeValue(0, scenario.attrType, "address.street"))
				} else {
					invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(0, scenario.attrType, "address.street"), scenario.attrValue)
				}

				afterAddress := srv.state.ResultSet().Items()[0]["address"].(*types.AttributeValueMemberM)
				after := afterAddress.Value["street"]

				assert.Equal(t, scenario.expected, after)
				assert.True(t, srv.state.ResultSet().IsDirty(0))
			})
		}
	})
}

func TestTableWriteController_DeleteAttribute(t *testing.T) {
	t.Run("should delete top level attribute", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		invokeCommand(t, srv.readController.Init())
		before, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("age")
		assert.Equal(t, "23", before)
		assert.False(t, srv.state.ResultSet().IsDirty(0))

		invokeCommand(t, srv.writeController.DeleteAttribute(0, "age"))

		_, hasAge := srv.state.ResultSet().Items()[0]["age"]
		assert.False(t, hasAge)
	})

	t.Run("should delete attribute of map", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		invokeCommand(t, srv.readController.Init())

		beforeAddress := srv.state.ResultSet().Items()[0]["address"].(*types.AttributeValueMemberM)
		beforeStreet := beforeAddress.Value["no"].(*types.AttributeValueMemberN).Value

		assert.Equal(t, "123", beforeStreet)
		assert.False(t, srv.state.ResultSet().IsDirty(0))

		invokeCommand(t, srv.writeController.DeleteAttribute(0, "address.no"))

		afterAddress := srv.state.ResultSet().Items()[0]["address"].(*types.AttributeValueMemberM)
		_, hasStreet := afterAddress.Value["no"]

		assert.False(t, hasStreet)
	})
}

func TestTableWriteController_PutItem(t *testing.T) {
	t.Run("should put the selected item if dirty", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		// Read the table
		invokeCommand(t, srv.readController.Init())
		before, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, srv.state.ResultSet().IsDirty(0))

		// Modify the item and put it
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandWithPrompt(t, srv.writeController.PutItems(), "y")

		// Rescan the table
		invokeCommand(t, srv.readController.Rescan())
		after, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "a new value", after)
		assert.False(t, srv.state.ResultSet().IsDirty(0))
	})

	t.Run("should not put the selected item if user does not confirm", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		// Read the table
		invokeCommand(t, srv.readController.Init())
		before, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, srv.state.ResultSet().IsDirty(0))

		// Modify the item but do not put it
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandWithPrompt(t, srv.writeController.PutItems(), "n")

		current, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "a new value", current)
		assert.True(t, srv.state.ResultSet().IsDirty(0))

		// Rescan the table to confirm item is not modified
		invokeCommandWithPrompt(t, srv.readController.Rescan(), "y")
		after, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", after)
		assert.False(t, srv.state.ResultSet().IsDirty(0))
	})

	t.Run("should not put the selected item if not dirty", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		// Read the table
		invokeCommand(t, srv.readController.Init())
		before, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, srv.state.ResultSet().IsDirty(0))

		invokeCommand(t, srv.writeController.PutItems())
	})

	t.Run("should not put the selected item if in read-only mode", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table", isReadOnly: true})

		// Read the table
		invokeCommand(t, srv.readController.Init())
		before, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, srv.state.ResultSet().IsDirty(0))

		// Modify the item but do not put it
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandExpectingError(t, srv.writeController.PutItems())

		current, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "a new value", current)
		assert.True(t, srv.state.ResultSet().IsDirty(0))

		// Rescan the table to confirm item is not modified
		invokeCommandWithPrompt(t, srv.readController.Rescan(), "y")
		after, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", after)
		assert.False(t, srv.state.ResultSet().IsDirty(0))
	})
}

func TestTableWriteController_PutItems(t *testing.T) {
	t.Run("should put all dirty items if none are marked", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		invokeCommand(t, srv.readController.Init())

		// Modify the item and put it
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(2, models.StringItemType, "alpha"), "another new value")

		invokeCommandWithPrompt(t, srv.writeController.PutItems(), "y")

		// Rescan the table
		invokeCommand(t, srv.readController.Rescan())

		assert.Equal(t, "a new value", srv.state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "another new value", srv.state.ResultSet().Items()[2]["alpha"].(*types.AttributeValueMemberS).Value)

		assert.False(t, srv.state.ResultSet().IsDirty(0))
		assert.False(t, srv.state.ResultSet().IsDirty(2))
	})

	t.Run("only put marked items", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		invokeCommand(t, srv.readController.Init())

		// Modify the item and put it
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(2, models.StringItemType, "alpha"), "another new value")
		invokeCommand(t, srv.writeController.ToggleMark(0))

		invokeCommandWithPrompt(t, srv.writeController.PutItems(), "y")

		// Verify dirty items are unchanged
		assert.Equal(t, "a new value", srv.state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "another new value", srv.state.ResultSet().Items()[2]["alpha"].(*types.AttributeValueMemberS).Value)

		assert.False(t, srv.state.ResultSet().IsDirty(0))
		assert.True(t, srv.state.ResultSet().IsDirty(2))

		// Rescan the table and verify dirty items were not written
		invokeCommandWithPrompt(t, srv.readController.Rescan(), "y")

		assert.Equal(t, "a new value", srv.state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Nil(t, srv.state.ResultSet().Items()[2]["alpha"])

		assert.False(t, srv.state.ResultSet().IsDirty(0))
		assert.False(t, srv.state.ResultSet().IsDirty(2))
	})

	t.Run("do not put marked items which are not dirty", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		invokeCommand(t, srv.readController.Init())

		// Modify the item and put it
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(2, models.StringItemType, "alpha"), "another new value")
		invokeCommand(t, srv.writeController.ToggleMark(1))

		invokeCommand(t, srv.writeController.PutItems())

		// Verify dirty items are unchanged
		assert.Equal(t, "a new value", srv.state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "another new value", srv.state.ResultSet().Items()[2]["alpha"].(*types.AttributeValueMemberS).Value)

		assert.True(t, srv.state.ResultSet().IsDirty(0))
		assert.True(t, srv.state.ResultSet().IsDirty(2))

		// Rescan the table and verify dirty items were not written
		invokeCommandWithPrompt(t, srv.readController.Rescan(), "y")

		assert.Equal(t, "This is some value", srv.state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Nil(t, srv.state.ResultSet().Items()[2]["alpha"])

		assert.False(t, srv.state.ResultSet().IsDirty(0))
		assert.False(t, srv.state.ResultSet().IsDirty(2))
	})

	t.Run("do nothing if in read-only mode", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table", isReadOnly: true})

		invokeCommand(t, srv.readController.Init())

		// Modify the item and put it
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(2, models.StringItemType, "alpha"), "another new value")
		invokeCommand(t, srv.writeController.ToggleMark(0))

		invokeCommandExpectingError(t, srv.writeController.PutItems())

		// Verify dirty items are unchanged
		assert.Equal(t, "a new value", srv.state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Equal(t, "another new value", srv.state.ResultSet().Items()[2]["alpha"].(*types.AttributeValueMemberS).Value)

		assert.True(t, srv.state.ResultSet().IsDirty(0))
		assert.True(t, srv.state.ResultSet().IsDirty(2))

		// Rescan the table and verify dirty items were not written
		invokeCommandWithPrompt(t, srv.readController.Rescan(), "y")

		assert.Equal(t, "This is some value", srv.state.ResultSet().Items()[0]["alpha"].(*types.AttributeValueMemberS).Value)
		assert.Nil(t, srv.state.ResultSet().Items()[2]["alpha"])

		assert.False(t, srv.state.ResultSet().IsDirty(0))
		assert.False(t, srv.state.ResultSet().IsDirty(2))
	})
}

func TestTableWriteController_TouchItem(t *testing.T) {
	t.Run("should put the selected item if unmodified", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		// Read the table
		invokeCommand(t, srv.readController.Init())
		before, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, srv.state.ResultSet().IsDirty(0))

		// Modify the item and put it
		invokeCommandWithPrompt(t, srv.writeController.TouchItem(0), "y")

		// Rescan the table
		invokeCommand(t, srv.readController.Rescan())
		after, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", after)
		assert.False(t, srv.state.ResultSet().IsDirty(0))
	})

	t.Run("should not put the selected item if modified", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		// Read the table
		invokeCommand(t, srv.readController.Init())
		before, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, srv.state.ResultSet().IsDirty(0))

		// Modify the item and put it
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandExpectingError(t, srv.writeController.TouchItem(0))
	})

	t.Run("should not put the selected item if in read-only mode", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table", isReadOnly: true})

		// Read the table
		invokeCommand(t, srv.readController.Init())
		before, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, srv.state.ResultSet().IsDirty(0))

		// Modify the item and put it
		invokeCommandExpectingError(t, srv.writeController.TouchItem(0))

		// Rescan the table
		invokeCommand(t, srv.readController.Rescan())
		after, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", after)
		assert.False(t, srv.state.ResultSet().IsDirty(0))
	})
}

func TestTableWriteController_NoisyTouchItem(t *testing.T) {
	t.Run("should delete and put the selected item if unmodified", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		// Read the table
		invokeCommand(t, srv.readController.Init())
		before, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, srv.state.ResultSet().IsDirty(0))

		// Modify the item and put it
		invokeCommandWithPrompt(t, srv.writeController.NoisyTouchItem(0), "y")

		// Rescan the table
		invokeCommand(t, srv.readController.Rescan())
		after, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", after)
		assert.False(t, srv.state.ResultSet().IsDirty(0))
	})

	t.Run("should not put the selected item if modified", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		// Read the table
		invokeCommand(t, srv.readController.Init())
		before, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, srv.state.ResultSet().IsDirty(0))

		// Modify the item and put it
		invokeCommandWithPrompt(t, srv.writeController.SetAttributeValue(0, models.StringItemType, "alpha"), "a new value")
		invokeCommandExpectingError(t, srv.writeController.NoisyTouchItem(0))
	})

	t.Run("should not put the selected item if in read-only mode", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table", isReadOnly: true})

		// Read the table
		invokeCommand(t, srv.readController.Init())
		before, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", before)
		assert.False(t, srv.state.ResultSet().IsDirty(0))

		// Modify the item and put it
		invokeCommandExpectingError(t, srv.writeController.NoisyTouchItem(0))
	})
}

func TestTableWriteController_DeleteMarked(t *testing.T) {
	t.Run("should delete marked items", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table"})

		// Read the table
		invokeCommand(t, srv.readController.Init())
		assert.Len(t, srv.state.ResultSet().Items(), 3)

		// Mark some items
		invokeCommand(t, srv.writeController.ToggleMark(0))
		invokeCommand(t, srv.writeController.ToggleMark(2))

		// Delete it
		invokeCommandWithPrompt(t, srv.writeController.DeleteMarked(), "y")

		// Rescan and confirm marked items are deleted
		invokeCommand(t, srv.readController.Init())
		assert.Len(t, srv.state.ResultSet().Items(), 1)
		after, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is another some value", after)
	})

	t.Run("should not delete marked items if in read-only mode", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "alpha-table", isReadOnly: true})

		// Read the table
		invokeCommand(t, srv.readController.Init())
		assert.Len(t, srv.state.ResultSet().Items(), 3)

		// Mark some items
		invokeCommand(t, srv.writeController.ToggleMark(0))
		invokeCommand(t, srv.writeController.ToggleMark(2))

		// Delete it
		invokeCommandExpectingError(t, srv.writeController.DeleteMarked())

		// Rescan and confirm marked items are not deleted
		invokeCommand(t, srv.readController.Init())
		assert.Len(t, srv.state.ResultSet().Items(), 3)
		after, _ := srv.state.ResultSet().Items()[0].AttributeValueAsString("alpha")
		assert.Equal(t, "This is some value", after)
	})
}

type services struct {
	state              *controllers.State
	settingProvider    controllers.SettingsProvider
	readController     *controllers.TableReadController
	writeController    *controllers.TableWriteController
	settingsController *controllers.SettingsController
	columnsController  *controllers.ColumnsController
	exportController   *controllers.ExportController
}

type serviceConfig struct {
	tableName  string
	isReadOnly bool
}

func newService(t *testing.T, cfg serviceConfig) *services {
	ws := testworkspace.New(t)

	resultSetSnapshotStore := workspacestore.NewResultSetSnapshotStore(ws)
	settingStore := settingstore.New(ws)
	workspaceService := workspaces_service.NewService(resultSetSnapshotStore)
	itemRendererService := itemrenderer.NewService(itemrenderer.PlainTextRenderer(), itemrenderer.PlainTextRenderer())

	client := testdynamo.SetupTestTable(t, testData)

	provider := dynamo.NewProvider(client)
	service := tables.NewService(provider, settingStore)
	eventBus := bus.New()

	state := controllers.NewState()
	jobsController := controllers.NewJobsController(jobs.NewService(eventBus), eventBus, true)
	readController := controllers.NewTableReadController(state, service, workspaceService, itemRendererService, jobsController, eventBus, cfg.tableName)
	writeController := controllers.NewTableWriteController(state, service, jobsController, readController, settingStore)
	settingsController := controllers.NewSettingsController(settingStore)
	columnsController := controllers.NewColumnsController(eventBus)
	exportController := controllers.NewExportController(state, columnsController)

	if cfg.isReadOnly {
		if err := settingStore.SetReadOnly(cfg.isReadOnly); err != nil {
			t.Errorf("cannot set ro: %v", err)
		}
	}

	return &services{
		state:              state,
		settingProvider:    settingStore,
		readController:     readController,
		writeController:    writeController,
		settingsController: settingsController,
		columnsController:  columnsController,
		exportController:   exportController,
	}
}
