package pluginruntime_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPluginRuntime_Ext_RegisterCommand(t *testing.T) {
	srv := setupTestService(t)

	pluginChan, err := srv.LoadScript("test.js", `
		const ext = require("audax:dynamo-browse/ext");

		ext.registerCommand("the-command", () => {
			console.log("the command");
		});
	`)
	assert.NoError(t, err)
	assert.NoError(t, (<-pluginChan).Err())

	cmd := srv.MissingCommand("the-command")
	assert.NotNil(t, cmd)
}
