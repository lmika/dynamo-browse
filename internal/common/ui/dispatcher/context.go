package dispatcher

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/common/ui/uimodels"
)

type dispatcherContext struct {
	d *Dispatcher
}

func (dc dispatcherContext) Messagef(format string, args ...interface{}) {
	dc.d.publisher.Send(events.Message(fmt.Sprintf(format, args...)))
}

func (dc dispatcherContext) Send(teaMessage tea.Msg) {
	dc.d.publisher.Send(teaMessage)
}

func (dc dispatcherContext) Message(msg string) {
	dc.d.publisher.Send(events.Message(msg))
}

func (dc dispatcherContext) Input(prompt string, onDone uimodels.Operation) {
	dc.d.publisher.Send(events.PromptForInput{
		OnDone: onDone,
	})
}
