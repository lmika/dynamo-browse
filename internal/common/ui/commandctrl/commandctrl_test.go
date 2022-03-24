package commandctrl_test

import (
	"context"
	"testing"

	"github.com/lmika/awstools/internal/common/ui/commandctrl"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/test/testuictx"
	"github.com/stretchr/testify/assert"
)

func TestCommandController_Prompt(t *testing.T) {
	t.Run("prompt user for a command", func(t *testing.T) {
		cmd := commandctrl.NewCommandController(nil)

		ctx, uiCtx := testuictx.New(context.Background())
		err := cmd.Prompt().Execute(ctx)

		assert.NoError(t, err)

		promptMsg, ok := uiCtx.Messages[0].(events.PromptForInput)
		assert.True(t, ok)
		assert.Equal(t, ":", promptMsg.Prompt)
	})
}
