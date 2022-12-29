package controllers_test

import (
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestScriptController_RunScript(t *testing.T) {
	t.Run("should execute scripts successfully", func(t *testing.T) {
		srv := newService(t, serviceConfig{
			scriptFS: testScriptFile(t, "test.tm", `
				ui.print("Hello world")
			`),
		})
		doneChan := make(chan error)

		msg := srv.scriptController.RunScript("test.tm", doneChan)
		assert.Nil(t, msg)
		assert.NoError(t, <-doneChan)

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
			doneChan := make(chan error)

			invokeCommand(t, srv.readController.Init())
			msg := srv.scriptController.RunScript("test.tm", doneChan)
			assert.Nil(t, msg)
			assert.NoError(t, <-doneChan)

			assert.Len(t, srv.msgSender.msgs, 1)
			assert.Equal(t, events.StatusMsg("3"), srv.msgSender.msgs[0])
		})
	})

	t.Run("session.query", func(t *testing.T) {
		t.Run("should run query against current table", func(t *testing.T) {
			srv := newService(t, serviceConfig{
				tableName: "alpha-table",
				scriptFS: testScriptFile(t, "test.tm", `
					rs := session.query('pk="abc"').unwrap()
					ui.print(rs.length)
				`),
			})
			doneChan := make(chan error)

			invokeCommand(t, srv.readController.Init())
			msg := srv.scriptController.RunScript("test.tm", doneChan)
			assert.Nil(t, msg)
			assert.NoError(t, <-doneChan)

			assert.Len(t, srv.msgSender.msgs, 1)
			assert.Equal(t, events.StatusMsg("2"), srv.msgSender.msgs[0])
		})
	})
}

func TestScriptController_LookupCommand(t *testing.T) {
	t.Run("should schedule the script on a separate go-routine", func(t *testing.T) {
		srv := newService(t, serviceConfig{
			tableName: "alpha-table",
			scriptFS: testScriptFile(t, "test.tm", `
				ext.command("mycommand", func(name) {
					ui.print("Hello, ", name)
				})
			`),
		})

		waitChan := srv.msgSender.afterNextMessage()

		invokeCommand(t, srv.scriptController.LoadScript("test.tm"))
		invokeCommand(t, srv.commandController.Execute(`mycommand "test name"`))

		select {
		case <-waitChan:
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout waiting for next message")
		}

		assert.Len(t, srv.msgSender.msgs, 1)
		assert.Equal(t, events.StatusMsg("Hello, test name"), srv.msgSender.msgs[0])
	})
}
