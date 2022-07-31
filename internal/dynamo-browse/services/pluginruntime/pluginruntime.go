package pluginruntime

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/pkg/errors"
	"os"
)

type Service struct {
	registry *require.Registry
}

func New() *Service {
	return &Service{
		registry: new(require.Registry),
	}
}

func (s *Service) Load(filename string) (*Plugin, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to load plugin %v", filename)
	}

	runtime := goja.New()
	s.registry.Enable(runtime)
	console.Enable(runtime)

	if _, err := runtime.RunScript(filename, string(f)); err != nil {
		return nil, errors.Wrapf(err, "error loading plugin %v", filename)
	}

	return &Plugin{}, nil
}
