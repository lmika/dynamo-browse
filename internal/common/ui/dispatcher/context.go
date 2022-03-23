package dispatcher

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/common/ui/uimodels"
)

type DispatcherContext struct {
	Publisher MessagePublisher
}

func (dc DispatcherContext) Messagef(format string, args ...interface{}) {
	dc.Publisher.Send(events.Message(fmt.Sprintf(format, args...)))
}

func (dc DispatcherContext) Send(teaMessage tea.Msg) {
	dc.Publisher.Send(teaMessage)
}

func (dc DispatcherContext) Message(msg string) {
	dc.Publisher.Send(events.Message(msg))
}

func (dc DispatcherContext) Input(prompt string, onDone uimodels.Operation) {
	dc.Publisher.Send(events.PromptForInput{
		Prompt: prompt,
		OnDone: onDone,
	})
}
