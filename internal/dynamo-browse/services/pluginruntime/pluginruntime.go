package pluginruntime

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/dop251/goja_nodejs/require"
	"github.com/lmika/audax/internal/common/ui/commandctrl"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/pkg/errors"
	"log"
	"os"
)

type Service struct {
	registry  *require.Registry
	eventLoop *eventloop.EventLoop
	state     *controllers.State

	userCommands map[string]goja.Callable

	msgSender func(msg tea.Msg)
}

func New(state *controllers.State) *Service {
	srv := &Service{
		state:        state,
		userCommands: make(map[string]goja.Callable),
	}

	srv.registry = new(require.Registry)
	srv.registry.RegisterNativeModule("audax:dynamo-browse", audaxDynamoBrowse(srv))

	srv.eventLoop = eventloop.NewEventLoop(eventloop.WithRegistry(srv.registry))

	return srv
}

func (s *Service) SetMessageSender(msgFn func(msg tea.Msg)) {
	s.msgSender = msgFn
}

func (s *Service) postMessage(msg tea.Msg) {
	if s.msgSender != nil {
		s.msgSender(msg)
	}
}

func (s *Service) Start() {
	s.eventLoop.Start()
}

func (s *Service) Load(filename string) (*Plugin, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to load plugin %v", filename)
	}

	pgrm, err := goja.Compile(filename, string(f), true)
	if err != nil {
		return nil, errors.Wrapf(err, "compile error %v", filename)
	}

	//rt := goja.New()
	//s.registry.Enable(rt)
	//console.Enable(rt)

	plugin := &Plugin{pgrm: pgrm}
	s.eventLoop.RunOnLoop(func(rt *goja.Runtime) {
		if err := plugin.Run(rt); err != nil {
			log.Printf("error: %v", err)
		}
	})

	return nil, nil
}

func (s *Service) MissingCommand(name string) commandctrl.Command {
	callable := s.userCommands[name]
	if callable == nil {
		return nil
	}

	return func(args []string) tea.Cmd {
		s.eventLoop.RunOnLoop(func(rt *goja.Runtime) {
			argValues := make([]goja.Value, len(args))
			for i, a := range args {
				argValues[i] = rt.ToValue(a)
			}

			// TODO: deal with error
			if _, err := callable(goja.Undefined(), argValues...); err != nil {
				log.Printf("error: %v", err)
			}
		})
		return nil
	}
}
