package commandctrl_test

import (
	"testing"

	"github.com/lmika/awstools/internal/common/ui/commandctrl"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/stretchr/testify/assert"
)

func TestCommandController_Prompt(t *testing.T) {
	t.Run("prompt user for a command", func(t *testing.T) {
		cmd := commandctrl.NewCommandController()

		res := cmd.Prompt()()

		promptForInputMsg, ok := res.(events.PromptForInputMsg)
		assert.True(t, ok)
		assert.Equal(t, ":", promptForInputMsg.Prompt)
	})
}
