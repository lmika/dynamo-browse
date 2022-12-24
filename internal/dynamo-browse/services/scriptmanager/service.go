package scriptmanager

import (
	"context"
	"github.com/cloudcmds/tamarin/exec"
	"github.com/cloudcmds/tamarin/scope"
	"github.com/pkg/errors"
	"io/fs"
)

type Service struct {
	fs     fs.FS
	ifaces Ifaces
}

func New(fs fs.FS) *Service {
	return &Service{
		fs: fs,
	}
}

func (s *Service) SetIFaces(ifaces Ifaces) {
	s.ifaces = ifaces
}

func (s *Service) LoadScript(filename string) error {
	return nil
}

func (s *Service) RunAdHocScript(ctx context.Context, filename string) chan error {
	errChan := make(chan error)
	go s.startAdHocScript(ctx, filename, errChan)
	return errChan
}

func (s *Service) StartAdHocScript(ctx context.Context, filename string, errChan chan error) {
	go s.startAdHocScript(ctx, filename, errChan)
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

	if _, err = exec.Execute(ctx, exec.Opts{
		Input: string(code),
		File:  filename,
		Scope: scp,
	}); err != nil {
		errChan <- errors.Wrapf(err, "script %v", filename)
		return
	}
}
