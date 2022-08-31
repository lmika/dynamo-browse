package pluginruntime

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/lmika/audax/internal/common/ui/events"
	"strings"
)

func audaxDynamoUI(service *Service) require.ModuleLoader {
	return func(rt *goja.Runtime, module *goja.Object) {
		o := module.Get("exports").(*goja.Object)

		// UI
		o.Set("print", func(call goja.FunctionCall) goja.Value {
			sb := new(strings.Builder)
			for _, arg := range call.Arguments {
				sb.WriteString(arg.String())
			}

			service.postMessage(events.StatusMsg(sb.String()))

			// post alert
			return goja.Undefined()
		})
		o.Set("prompt", func(msg string) goja.Value {
			p, resolve, _ := rt.NewPromise()
			service.postMessage(events.PromptForInputMsg{
				Prompt: msg,
				OnDone: func(value string) tea.Msg {
					service.eventLoop.RunOnLoop(func(rt *goja.Runtime) {
						resolve(value)
					})
					return nil
				},
			})
			return rt.ToValue(p)
		})
	}
}
