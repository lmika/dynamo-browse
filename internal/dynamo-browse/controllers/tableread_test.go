package controllers_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/controllers"
	"github.com/lmika/dynamo-browse/test/testdynamo"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestTableReadController_InitTable(t *testing.T) {
	t.Run("should prompt for table if no table name provided", func(t *testing.T) {
		srv := newService(t, serviceConfig{})

		event := srv.readController.Init()

		assert.IsType(t, controllers.PromptForTableMsg{}, event)
	})

	t.Run("should scan table if table name provided", func(t *testing.T) {
		srv := newService(t, serviceConfig{})

		event := srv.readController.Init()

		assert.IsType(t, controllers.PromptForTableMsg{}, event)
	})
}

func TestTableReadController_ListTables(t *testing.T) {
	t.Run("returns a list of tables", func(t *testing.T) {
		srv := newService(t, serviceConfig{})

		event := srv.readController.ListTables(false).(controllers.PromptForTableMsg)

		assert.Equal(t, []string{"alpha-table", "bravo-table", "count-to-30"}, event.Tables)

		selectedEvent := event.OnSelected("alpha-table")

		resultSet := selectedEvent.(controllers.NewResultSet)
		assert.Equal(t, "alpha-table", resultSet.ResultSet.TableInfo.Name)
		assert.Equal(t, "pk", resultSet.ResultSet.TableInfo.Keys.PartitionKey)
		assert.Equal(t, "sk", resultSet.ResultSet.TableInfo.Keys.SortKey)
	})
}

func TestTableReadController_Rescan(t *testing.T) {
	t.Run("should perform a rescan", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "bravo-table"})

		invokeCommand(t, srv.readController.Init())
		invokeCommand(t, srv.readController.Rescan())
	})

	t.Run("should prompt to rescan if any dirty rows", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "bravo-table"})

		invokeCommand(t, srv.readController.Init())

		srv.state.ResultSet().SetDirty(0, true)

		invokeCommandWithPrompt(t, srv.readController.Rescan(), "y")

		assert.False(t, srv.state.ResultSet().IsDirty(0))
	})

	t.Run("should not rescan if any dirty rows", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "bravo-table"})

		invokeCommand(t, srv.readController.Init())

		srv.state.ResultSet().SetDirty(0, true)

		invokeCommandWithPrompt(t, srv.readController.Rescan(), "n")

		assert.True(t, srv.state.ResultSet().IsDirty(0))
	})
}

func TestTableReadController_Query(t *testing.T) {
	t.Run("should run scan with filter based on user query", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "bravo-table"})

		tempFile := tempFile(t)

		invokeCommand(t, srv.readController.Init())
		invokeCommandWithPrompts(t, srv.readController.PromptForQuery(), `pk ^= "abc"`)
		invokeCommand(t, srv.exportController.ExportCSV(tempFile, controllers.ExportOptions{}))

		bts, err := os.ReadFile(tempFile)
		assert.NoError(t, err)

		assert.Equal(t, string(bts), strings.Join([]string{
			"pk,sk,alpha,beta\n",
			"abc,222,This is another some value,1231\n",
		}, ""))
	})

	t.Run("should run query on index with filter based on user query", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "bravo-table"})

		tempFile := tempFile(t)

		invokeCommand(t, srv.readController.Init())
		invokeCommandWithPrompts(t, srv.readController.PromptForQuery(), `alpha = "This is some value"`)
		invokeCommand(t, srv.exportController.ExportCSV(tempFile, controllers.ExportOptions{}))

		bts, err := os.ReadFile(tempFile)
		assert.NoError(t, err)

		assert.Equal(t, string(bts), strings.Join([]string{
			"pk,sk,alpha\n",
			"foo,bar,This is some value\n",
		}, ""))
	})

	t.Run("should return error if result set is not set", func(t *testing.T) {
		srv := newService(t, serviceConfig{tableName: "non-existant-table"})

		tempFile := tempFile(t)

		invokeCommandExpectingError(t, srv.readController.Init())
		invokeCommandExpectingError(t, srv.exportController.ExportCSV(tempFile, controllers.ExportOptions{}))
	})
}

func TestTableReadController_NextPage(t *testing.T) {
	t.Run("should return successive pages of results", func(t *testing.T) {
		scenarios := []struct {
			pageLimit         int
			expectedPageSizes []int
		}{
			{pageLimit: 10, expectedPageSizes: []int{10, 10, 10}},
			{pageLimit: 5, expectedPageSizes: []int{5, 5, 5, 5, 5, 5}},
			{pageLimit: 13, expectedPageSizes: []int{13, 13, 4}},
			{pageLimit: 7, expectedPageSizes: []int{7, 7, 7, 7, 2}},
			{pageLimit: 3, expectedPageSizes: []int{3, 3, 3, 3, 3, 3, 3, 3, 3, 3}},
			{pageLimit: 50, expectedPageSizes: []int{30}},
			{pageLimit: 100, expectedPageSizes: []int{30}},
		}

		for _, scenario := range scenarios {
			t.Run(fmt.Sprintf("page size = %v", scenario.pageLimit), func(t *testing.T) {
				srv := newService(t, serviceConfig{tableName: "count-to-30", defaultLimit: scenario.pageLimit})

				invokeCommand(t, srv.readController.Init())

				var currentCount = 1

				// Go through each page and confirm that the items are correct
				for i, pageSize := range scenario.expectedPageSizes {
					if i > 0 {
						invokeCommand(t, srv.readController.NextPage())
					}

					rs := srv.state.ResultSet()
					assert.Len(t, rs.Items(), pageSize)
					for _, item := range rs.Items() {
						assert.Equal(t, fmt.Sprintf("NUM#%02d", currentCount), item["sk"].(*types.AttributeValueMemberS).Value)
						currentCount += 1
					}
				}

				// Attempt to get the last page of results, but no more results.  This should not
				// clear the current page of results
				invokeCommand(t, srv.readController.NextPage())

				rs := srv.state.ResultSet()
				assert.Len(t, rs.Items(), scenario.expectedPageSizes[len(scenario.expectedPageSizes)-1])
			})
		}
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
		Index:     []string{"alpha"},
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
	{
		TableName: "count-to-30",
		Data: sequenceToN(1, 30, func(n int) map[string]any {
			return map[string]any{
				"pk":  "NUM",
				"sk":  fmt.Sprintf("NUM#%02d", n),
				"num": n,
			}
		}),
	},
}

func sequenceToN[T any](from int, to int, fn func(n int) T) []T {
	ns := make([]T, 0, to-from+1)
	for i := from; i <= to; i++ {
		ns = append(ns, fn(i))
	}
	return ns
}
