package pluginruntime

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"log"
)

func audaxDynamoExt(service *Service) require.ModuleLoader {
	return func(rt *goja.Runtime, module *goja.Object) {
		o := module.Get("exports").(*goja.Object)

		// Session
		o.Set("registerCommand", func(name string, fn goja.Callable) {
			log.Printf("registering user command: %v", name)
			service.userCommands[name] = fn
		})
	}
}
