package scriptmanager

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/keybindings"
	"github.com/pkg/errors"
	"github.com/risor-io/risor"
	"github.com/risor-io/risor/object"
)

type Service struct {
	lookupPaths []fs.FS
	ifaces      Ifaces
	options     Options
	sched       *scriptScheduler
	plugins     []*ScriptPlugin
}

func New(opts ...ServiceOption) *Service {
	srv := &Service{
		lookupPaths: nil,
		sched:       newScriptScheduler(),
	}
	for _, opt := range opts {
		opt(srv)
	}
	return srv
}

func (s *Service) SetLookupPaths(fs []fs.FS) {
	s.lookupPaths = fs
}

func (s *Service) SetDefaultOptions(options Options) {
	s.options = options
}

func (s *Service) SetIFaces(ifaces Ifaces) {
	s.ifaces = ifaces
}

func (s *Service) LoadScript(ctx context.Context, filename string) (*ScriptPlugin, error) {
	resChan := make(chan loadedScriptResult)

	if err := s.sched.startJobOnceFree(ctx, func(ctx context.Context) {
		s.loadScript(ctx, filename, resChan)
	}); err != nil {
		return nil, err
	}

	res := <-resChan
	if res.err != nil {
		return nil, res.err
	}

	// Look for the previous version.  If one is there, replace it, otherwise add it
	// TODO: this should probably be protected by a mutex
	newPlugin := res.scriptPlugin
	for i, p := range s.plugins {
		if p.name == newPlugin.name {
			s.plugins[i] = newPlugin
			return newPlugin, nil
		}
	}

	s.plugins = append(s.plugins, newPlugin)
	return newPlugin, nil
}

func (s *Service) RunAdHocScript(ctx context.Context, filename string) chan error {
	errChan := make(chan error)
	go s.startAdHocScript(ctx, filename, errChan)
	return errChan
}

func (s *Service) StartAdHocScript(ctx context.Context, filename string, errChan chan error) error {
	return s.sched.startJobOnceFree(ctx, func(ctx context.Context) {
		s.startAdHocScript(ctx, filename, errChan)
	})
}

func (s *Service) startAdHocScript(ctx context.Context, filename string, errChan chan error) {
	defer close(errChan)

	code, err := s.readScript(filename, true)
	if err != nil {
		errChan <- errors.Wrapf(err, "cannot load script file %v", filename)
		return
	}

	ctx = ctxWithScriptEnv(ctx, scriptEnv{filename: filepath.Base(filename), options: s.options})

	if _, err := risor.Eval(ctx, code,
		risor.WithGlobals(s.builtins()),
		// risor.WithDefaultBuiltins(),
		// risor.WithDefaultModules(),
		// risor.WithBuiltins(s.builtins()),
	); err != nil {
		errChan <- errors.Wrapf(err, "script %v", filename)
		return
	}
}

type loadedScriptResult struct {
	scriptPlugin *ScriptPlugin
	err          error
}

func (s *Service) loadScript(ctx context.Context, filename string, resChan chan loadedScriptResult) {
	defer close(resChan)

	code, err := s.readScript(filename, false)
	if err != nil {
		resChan <- loadedScriptResult{err: errors.Wrapf(err, "cannot load script file %v", filename)}
		return
	}

	newPlugin := &ScriptPlugin{
		name:          strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)),
		scriptService: s,
	}

	ctx = ctxWithScriptEnv(ctx, scriptEnv{filename: filepath.Base(filename), options: s.options})

	if _, err := risor.Eval(ctx, code,
		// risor.WithDefaultBuiltins(),
		// risor.WithDefaultModules(),
		// risor.WithBuiltins(s.builtins()),
		risor.WithGlobals(s.builtins()),
		risor.WithGlobals(map[string]any{
			"ext": (&extModule{scriptPlugin: newPlugin}).register(),
		}),
	); err != nil {
		resChan <- loadedScriptResult{err: errors.Wrapf(err, "script %v", filename)}
		return
	}

	resChan <- loadedScriptResult{scriptPlugin: newPlugin}
}

func (s *Service) readScript(filename string, allowCwd bool) (string, error) {
	if allowCwd {
		if cwd, err := os.Getwd(); err == nil {
			fullScriptPath := filepath.Join(cwd, filename)
			log.Printf("checking %v", fullScriptPath)
			if stat, err := os.Stat(fullScriptPath); err == nil && !stat.IsDir() {
				code, err := os.ReadFile(filename)
				if err != nil {
					return "", err
				}
				return string(code), nil
			}
		} else {
			log.Printf("warn: cannot get cwd for reading script %v: %v", filename, err)
		}
	}

	if strings.HasPrefix(filename, string(filepath.Separator)) {
		code, err := os.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return string(code), nil
	}

	for _, currFS := range s.lookupPaths {
		log.Printf("checking %v/%v", currFS, filename)
		stat, err := fs.Stat(currFS, filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return "", err
			}
		} else if stat.IsDir() {
			continue
		}

		code, err := fs.ReadFile(currFS, filename)
		if err == nil {
			return string(code), nil
		} else {
			return "", err
		}
	}

	return "", os.ErrNotExist
}

// LookupCommand looks up a command defined by a script.
// TODO: Command should probably accept/return a chan error to indicate that this will run in a separate goroutine
func (s *Service) LookupCommand(name string) *Command {
	for _, p := range s.plugins {
		if cmd, hasCmd := p.definedCommands[name]; hasCmd {
			return cmd
		}
	}
	return nil
}

func (s *Service) LookupKeyBinding(key string) (string, *Command) {
	for _, p := range s.plugins {
		if bindingName, hasBinding := p.keyToKeyBinding[key]; hasBinding {
			if cmd, hasCmd := p.definedKeyBindings[bindingName]; hasCmd {
				return bindingName, cmd
			}
		}
	}
	return "", nil
}

func (s *Service) UnbindKey(key string) {
	for _, p := range s.plugins {
		if _, hasBinding := p.keyToKeyBinding[key]; hasBinding {
			delete(p.keyToKeyBinding, key)
		}
	}
}

func (s *Service) RebindKeyBinding(keyBinding string, newKey string) error {
	if newKey == "" {
		for _, p := range s.plugins {
			for k, b := range p.keyToKeyBinding {
				if b == keyBinding {
					delete(p.keyToKeyBinding, k)
				}
			}
		}
		return nil
	}

	for _, p := range s.plugins {
		if _, hasCmd := p.definedKeyBindings[keyBinding]; hasCmd {
			if newKey != "" {
				p.keyToKeyBinding[newKey] = keyBinding
			}
			return nil
		}
	}

	return keybindings.InvalidBindingError(keyBinding)
}

func (s *Service) builtins() map[string]any {
	return map[string]any{
		"ui":      (&uiModule{uiService: s.ifaces.UI}).register(),
		"session": (&sessionModule{sessionService: s.ifaces.Session}).register(),
		"os":      (&osModule{}).register(),
		"print":   object.NewBuiltin("print", printBuiltin),
		"printf":  object.NewBuiltin("printf", printfBuiltin),
	}
}
