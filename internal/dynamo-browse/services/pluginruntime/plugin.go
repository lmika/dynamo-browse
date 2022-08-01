package pluginruntime

import (
	"github.com/dop251/goja"
)

type Plugin struct {
	pgrm *goja.Program
}

func (p *Plugin) Run(rt *goja.Runtime) error {
	if _, err := rt.RunProgram(p.pgrm); err != nil {
		return err
	}
	return nil
}
