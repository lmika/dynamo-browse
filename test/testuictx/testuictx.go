package testuictx

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/common/ui/dispatcher"
	"github.com/lmika/dynamo-browse/internal/common/ui/uimodels"
)

func New(ctx context.Context) (context.Context, *TestUIContext) {
	td := &TestUIContext{}
	return uimodels.WithContext(ctx, dispatcher.DispatcherContext{td}), td
}

type TestUIContext struct {
	Messages []tea.Msg
}

func (t *TestUIContext) Send(msg tea.Msg) {
	t.Messages = append(t.Messages, msg)
}
