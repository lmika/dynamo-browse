package pluginruntime

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/queryexpr"
	"github.com/pkg/errors"
	"log"
	"strings"
)

func audaxDynamoBrowse(service *Service) require.ModuleLoader {
	return func(rt *goja.Runtime, module *goja.Object) {
		o := module.Get("exports").(*goja.Object)

		// Session
		sessionObj := rt.NewObject()
		sessionObj.Set("registerCommand", func(name string, fn goja.Callable) {
			log.Printf("registering user command: %v", name)
			service.userCommands[name] = fn
		})
		sessionObj.DefineAccessorProperty("currentResultSet", rt.ToValue(func() goja.Value {
			return newJSResultSet(rt, service.state.ResultSet())
		}), rt.ToValue(func(v goja.Value) error {
			obj := v.ToObject(rt)
			if obj == nil {
				return errors.Errorf("expected type to be resultSet")
			}

			do, isDo := obj.GetSymbol(resultSetSymbol).Export().(goja.DynamicObject)
			if !isDo {
				return errors.Errorf("expected type to be resultSet")
			}

			rsp, isRsp := do.(goProxyValue)
			if !isRsp {
				return errors.New("expected type to be resultSet")
			}
			resultSet := rsp.v.(*models.ResultSet)

			service.state.SetResultSetAndFilter(resultSet, "")
			service.postMessage(controllers.NewResultSet{ResultSet: resultSet})
			return nil
		}), goja.FLAG_FALSE, goja.FLAG_FALSE)
		sessionObj.Set("query", func(exprStr string, opts *goja.Object) goja.Value {
			currentResultSet := service.state.ResultSet()

			p, resolve, reject := rt.NewPromise()
			go func() {
				ctx := context.Background()

				expr, err := queryexpr.Parse(exprStr)
				if err != nil {
					reject(err)
					return
				}

				tableInfo := currentResultSet.TableInfo
				if opts != nil {
					if tableName, isStr := opts.Get("table").Export().(string); isStr && tableName != "" {
						if t, err := service.tableService.Describe(ctx, tableName); err == nil {
							tableInfo = t
						} else {
							reject(err)
							return
						}
					}
				}

				rs, err := service.tableService.ScanOrQuery(ctx, tableInfo, expr)
				if err != nil {
					reject(err)
					return
				}

				resolve(newJSResultSet(rt, rs))
			}()

			return rt.ToValue(p)
		})

		// UI
		uiObj := rt.NewObject()
		uiObj.Set("alert", func(call goja.FunctionCall) goja.Value {
			sb := new(strings.Builder)
			for _, arg := range call.Arguments {
				sb.WriteString(arg.String())
			}

			service.postMessage(events.StatusMsg(sb.String()))

			// post alert
			return goja.Undefined()
		})
		uiObj.Set("prompt", func(msg string) goja.Value {
			p, resolve, _ := rt.NewPromise()
			service.postMessage(events.PromptForInputMsg{
				Prompt: msg,
				OnDone: func(value string) tea.Cmd {
					service.eventLoop.RunOnLoop(func(rt *goja.Runtime) {
						resolve(value)
					})
					return nil
				},
			})
			return rt.ToValue(p)
		})

		o.Set("ui", uiObj)
		o.Set("session", sessionObj)
	}
}
