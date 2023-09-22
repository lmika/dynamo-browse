package controllers_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/controllers"
	"github.com/stretchr/testify/assert"
)

func TestScriptController_RunScript(t *testing.T) {
	t.Run("should execute scripts successfully", func(t *testing.T) {
		srv := newService(t, serviceConfig{
			scriptFS: testScriptFile(t, "test.tm", `
				ui.print("Hello world")
			`),
		})

		msg := srv.scriptController.RunScript("test.tm")
		assert.Nil(t, msg)

		srv.msgSender.waitForAtLeastOneMessages(t, 5*time.Second)

		assert.Len(t, srv.msgSender.msgs, 1)
		assert.Equal(t, events.StatusMsg("Hello world"), srv.msgSender.msgs[0])
	})

	t.Run("session.result_set", func(t *testing.T) {
		t.Run("should return current result set if not-nil", func(t *testing.T) {
			srv := newService(t, serviceConfig{
				tableName: "alpha-table",
				scriptFS: testScriptFile(t, "test.tm", `
					rs := session.result_set()
					ui.print(rs.length)
				`),
			})

			invokeCommand(t, srv.readController.Init())

			msg := srv.scriptController.RunScript("test.tm")
			assert.Nil(t, msg)

			srv.msgSender.waitForAtLeastOneMessages(t, 5*time.Second)

			assert.Len(t, srv.msgSender.msgs, 1)
			assert.Equal(t, events.StatusMsg("3"), srv.msgSender.msgs[0])
		})
	})

	t.Run("session.query", func(t *testing.T) {
		t.Run("should run query against current table", func(t *testing.T) {
			srv := newService(t, serviceConfig{
				tableName: "alpha-table",
				scriptFS: testScriptFile(t, "test.tm", `
					rs := session.query('pk="abc"')
					ui.print(rs.length)
				`),
			})

			invokeCommand(t, srv.readController.Init())
			msg := srv.scriptController.RunScript("test.tm")
			assert.Nil(t, msg)

			srv.msgSender.waitForAtLeastOneMessages(t, 5*time.Second)

			assert.Len(t, srv.msgSender.msgs, 1)
			assert.Equal(t, events.StatusMsg("2"), srv.msgSender.msgs[0])
		})

		t.Run("should run query against another table", func(t *testing.T) {
			srv := newService(t, serviceConfig{
				tableName: "alpha-table",
				scriptFS: testScriptFile(t, "test.tm", `
					rs := session.query('pk!="abc"', { table: "count-to-30" })
					ui.print(rs.length)
				`),
			})

			invokeCommand(t, srv.readController.Init())
			msg := srv.scriptController.RunScript("test.tm")
			assert.Nil(t, msg)

			srv.msgSender.waitForAtLeastOneMessages(t, 5*time.Second)

			assert.Len(t, srv.msgSender.msgs, 1)
			assert.Equal(t, events.StatusMsg("30"), srv.msgSender.msgs[0])
		})
	})

	t.Run("session.set_result_set", func(t *testing.T) {
		t.Run("should set the result set from the result of a query", func(t *testing.T) {
			srv := newService(t, serviceConfig{
				tableName: "alpha-table",
				scriptFS: testScriptFile(t, "test.tm", `
					rs := session.query('pk="abc"')
					session.set_result_set(rs)
				`),
			})

			invokeCommand(t, srv.readController.Init())
			msg := srv.scriptController.RunScript("test.tm")
			assert.Nil(t, msg)

			srv.msgSender.waitForAtLeastOneMessages(t, 5*time.Second)

			assert.Len(t, srv.msgSender.msgs, 1)
			assert.IsType(t, controllers.NewResultSet{}, srv.msgSender.msgs[0])
		})

		t.Run("changed attributes of the result set should show up as modified", func(t *testing.T) {
			srv := newService(t, serviceConfig{
				tableName: "alpha-table",
				scriptFS: testScriptFile(t, "test.tm", `
						rs := session.query('pk="abc"')
						rs[0].set_attr("pk", "131")
						session.set_result_set(rs)
					`),
			})

			invokeCommand(t, srv.readController.Init())
			msg := srv.scriptController.RunScript("test.tm")
			assert.Nil(t, msg)

			srv.msgSender.waitForAtLeastOneMessages(t, 5*time.Second)

			assert.Len(t, srv.msgSender.msgs, 1)
			assert.IsType(t, controllers.NewResultSet{}, srv.msgSender.msgs[0])

			assert.Equal(t, "131", srv.state.ResultSet().Items()[0]["pk"].(*types.AttributeValueMemberS).Value)
			assert.True(t, srv.state.ResultSet().IsDirty(0))
		})
	})
}

func TestScriptController_LookupCommand(t *testing.T) {
	t.Run("should schedule the script on a separate go-routine", func(t *testing.T) {
		scenarios := []struct {
			descr          string
			command        string
			expectedOutput string
		}{
			{descr: "command with arg", command: "mycommand \"test name\"", expectedOutput: "Hello, test name"},
			{descr: "command no arg", command: "mycommand", expectedOutput: "Hello, nil value"},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.descr, func(t *testing.T) {
				srv := newService(t, serviceConfig{
					tableName: "alpha-table",
					scriptFS: testScriptFile(t, "test.tm", `
						ext.command("mycommand", func(name = "nil value") {
							ui.print(sprintf("Hello, %v", name))
						})
					`),
				})

				invokeCommand(t, srv.scriptController.LoadScript("test.tm"))
				invokeCommand(t, srv.commandController.Execute(scenario.command))

				srv.msgSender.waitForAtLeastOneMessages(t, 5*time.Second)

				assert.Len(t, srv.msgSender.msgs, 1)
				assert.Equal(t, events.StatusMsg(scenario.expectedOutput), srv.msgSender.msgs[0])
			})
		}
	})

	t.Run("should only allow one script to run at a time", func(t *testing.T) {
		srv := newService(t, serviceConfig{
			tableName: "alpha-table",
			scriptFS: testScriptFile(t, "test.tm", `
				ext.command("mycommand", func() {
					time.sleep(1.5)
					ui.print("Done my thing")
				})
			`),
		})

		invokeCommand(t, srv.scriptController.LoadScript("test.tm"))

		invokeCommand(t, srv.commandController.Execute(`mycommand`))
		invokeCommandExpectingError(t, srv.commandController.Execute(`mycommand`))

		srv.msgSender.waitForAtLeastOneMessages(t, 5*time.Second)

		assert.Len(t, srv.msgSender.msgs, 1)
		assert.Equal(t, events.StatusMsg("Done my thing"), srv.msgSender.msgs[0])
	})

}
