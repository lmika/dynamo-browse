package pluginruntime

import (
	"context"
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"os/exec"
	"runtime"
)

func jsExecModule() require.ModuleLoader {
	return func(rt *goja.Runtime, module *goja.Object) {
		o := module.Get("exports").(*goja.Object)

		// This is non-standard and subject to change.  Really!
		o.Set("system", func(name string, args []string) goja.Value {
			p, resolve, reject := rt.NewPromise()
			go func() {
				ctx := context.Background()

				cmd := exec.CommandContext(ctx, name, args...)
				res, err := cmd.Output()
				if err != nil {
					reject(err)
					return
				}

				resolve(rt.ToValue(string(res)))
			}()

			return rt.ToValue(p)
		})
		o.Set("openUri", func(name string) error {
			return openURI(name)
		})
	}
}

/*
This is a modified copy of https://github.com/tebeka/desktop as written by Miki Tebeka <miki.tebeka@gmail.com>
*/
var commands = map[string][]string{
	"windows": []string{"cmd", "/c", "start"},
	"darwin":  []string{"open"},
	"linux":   []string{"xdg-open"},
}

// Open calls the OS default program for uri
// e.g. Open("http://www.google.com") will open the default browser on www.google.com
func openURI(uri string) error {
	run, ok := commands[runtime.GOOS]
	if !ok {
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}

	run = append(run, uri)
	cmd := exec.Command(run[0], run[1:]...)

	return cmd.Start()
}
