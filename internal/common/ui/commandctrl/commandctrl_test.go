package commandctrl_test

import (
	"github.com/lmika/audax/internal/common/ui/events"
	"testing"

	"github.com/lmika/audax/internal/common/ui/commandctrl"
	"github.com/stretchr/testify/assert"
)

func TestCommandController_Prompt(t *testing.T) {
	t.Run("prompt user for a command", func(t *testing.T) {
		cmd := commandctrl.NewCommandController()

		res := cmd.Prompt()

		promptForInputMsg, ok := res.(events.PromptForInputMsg)
		assert.True(t, ok)
		assert.Equal(t, ":", promptForInputMsg.Prompt)
	})
}
