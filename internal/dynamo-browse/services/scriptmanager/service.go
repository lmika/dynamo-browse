package scriptmanager

import (
	"context"
	"github.com/cloudcmds/tamarin/exec"
	"github.com/cloudcmds/tamarin/scope"
	"github.com/pkg/errors"
	"io/fs"
	"path/filepath"
)

type Service struct {
	fs      fs.FS
	ifaces  Ifaces
	sched   *scriptScheduler
	plugins []*ScriptPlugin
}

func New(fs fs.FS) *Service {
	return &Service{
		fs:    fs,
		sched: newScriptScheduler(),
	}
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

	code, err := fs.ReadFile(s.fs, filename)
	if err != nil {
		errChan <- errors.Wrapf(err, "cannot load script file %v", filename)
		return
	}

	// TODO: this should probably be a single scope with registered modules
	scp := scope.New(scope.Opts{})
	(&uiModule{uiService: s.ifaces.UI}).register(scp)
	(&sessionModule{sessionService: s.ifaces.Session}).register(scp)

	if _, err = exec.Execute(ctx, exec.Opts{
		Input: string(code),
		File:  filename,
		Scope: scp,
	}); err != nil {
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

	code, err := fs.ReadFile(s.fs, filename)
	if err != nil {
		resChan <- loadedScriptResult{err: errors.Wrapf(err, "cannot load script file %v", filename)}
		return
	}

	newPlugin := &ScriptPlugin{
		name:          filepath.Base(filename),
		scriptService: s,
	}

	// TODO: this should probably be a single scope with registered modules
	scp := scope.New(scope.Opts{})
	(&uiModule{uiService: s.ifaces.UI}).register(scp)
	(&sessionModule{sessionService: s.ifaces.Session}).register(scp)
	// END TODO

	(&extModule{scriptPlugin: newPlugin}).register(scp)

	if _, err = exec.Execute(ctx, exec.Opts{
		Input: string(code),
		File:  filename,
		Scope: scp,
	}); err != nil {
		resChan <- loadedScriptResult{err: errors.Wrapf(err, "script %v", filename)}
		return
	}

	resChan <- loadedScriptResult{scriptPlugin: newPlugin}
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
