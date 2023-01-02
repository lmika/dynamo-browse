package scriptmanager

import "io/fs"

type ServiceOption func(srv *Service)

func WithFS(fs ...fs.FS) ServiceOption {
	return func(srv *Service) {
		srv.lookupPaths = fs
	}
}
