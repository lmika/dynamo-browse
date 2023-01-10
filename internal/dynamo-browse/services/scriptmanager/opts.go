package scriptmanager

import (
	"context"
	"os"
)

type Options struct {
	// OSExecShell is the shell to use for calls to 'os.exec'.  If not defined,
	// it will use the value of the SHELL environment variable, otherwise it will
	// default to '/bin/bash'
	OSExecShell string

	// Permissions are the permissions the script can execute in
	Permissions Permissions
}

func (opts Options) configuredShell() string {
	if opts.OSExecShell != "" {
		return opts.OSExecShell
	}
	if shell, hasShell := os.LookupEnv("SHELL"); hasShell {
		return shell
	}
	return "/bin/bash"
}

// Permissions control the set of permissions of a script
type Permissions struct {
	// AllowShellCommands determines whether or not a script can execute shell commands.
	AllowShellCommands bool
}

type optionCtxKeyType struct{}

var optionCtxKey = optionCtxKeyType{}

func optionFromCtx(ctx context.Context) Options {
	perms, _ := ctx.Value(optionCtxKey).(Options)
	return perms
}

func ctxWithOptions(ctx context.Context, perms Options) context.Context {
	return context.WithValue(ctx, optionCtxKey, perms)
}
