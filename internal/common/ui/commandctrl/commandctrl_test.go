package commandctrl_test

import (
	"context"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services"
	"testing"

	"github.com/lmika/dynamo-browse/internal/common/ui/commandctrl"
	"github.com/stretchr/testify/assert"
)

func TestCommandController_Prompt(t *testing.T) {
	t.Run("prompt user for a command", func(t *testing.T) {
		cmd := commandctrl.NewCommandController(mockIterProvider{})

		res := cmd.Prompt()

		promptForInputMsg, ok := res.(events.PromptForInputMsg)
		assert.True(t, ok)
		assert.Equal(t, ":", promptForInputMsg.Prompt)
	})
}

type mockIterProvider struct {
}

func (m mockIterProvider) Iter(ctx context.Context, category string) services.HistoryProvider {
	return nil
}
