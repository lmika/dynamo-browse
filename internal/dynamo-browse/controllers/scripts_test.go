package controllers_test

import (
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/stretchr/testify/assert"
	"testing"
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
}
