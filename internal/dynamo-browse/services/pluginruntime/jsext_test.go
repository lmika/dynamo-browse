package pluginruntime_test

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/commandctrl"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPluginRuntime_Ext_RegisterCommand(t *testing.T) {
	t.Run("should register command", func(t *testing.T) {
		msgs := make(chan tea.Msg, 10)
		srv := setupTestService(t, msgs)

		pluginChan, err := srv.LoadScript("test.js", `
			const ext = require("audax:dynamo-browse/ext");
			const ui  = require("audax:dynamo-browse/ui");
	
			ext.registerCommand("the-command", () => {
				ui.print("the command");
			});
		`)
		assert.NoError(t, err)
		assert.NoError(t, (<-pluginChan).Err())

		cmd := srv.MissingCommand("the-command")
		assert.NotNil(t, cmd)

		cmd(commandctrl.ExecContext{}, []string{})
		assert.Equal(t, events.StatusMsg("the command"), <-msgs)
	})

	t.Run("command should accept arguments", func(t *testing.T) {
		msgs := make(chan tea.Msg, 10)
		srv := setupTestService(t, msgs)

		pluginChan, err := srv.LoadScript("test.js", `
			const ext = require("audax:dynamo-browse/ext");
			const ui  = require("audax:dynamo-browse/ui");
	
			ext.registerCommand("the-command", (x, y) => {
				ui.print("concat = [" + x + "," + y + "]");
			});
		`)
		assert.NoError(t, err)
		assert.NoError(t, (<-pluginChan).Err())

		cmd := srv.MissingCommand("the-command")
		assert.NotNil(t, cmd)

		cmd(commandctrl.ExecContext{}, []string{"abc", "123"})
		assert.Equal(t, events.StatusMsg("concat = [abc,123]"), <-msgs)
	})
}
