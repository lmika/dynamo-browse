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

	// AllowEnv determines whether or not a script can access environment variables
	AllowEnv bool
}

// scriptEnv is the runtime environment for a particular script execution
type scriptEnv struct {
	filename string
	options  Options
}

type scriptEnvKeyType struct{}

var scriptEnvKey = scriptEnvKeyType{}

func scriptEnvFromCtx(ctx context.Context) scriptEnv {
	perms, _ := ctx.Value(scriptEnvKey).(scriptEnv)
	return perms
}

func ctxWithScriptEnv(ctx context.Context, perms scriptEnv) context.Context {
	return context.WithValue(ctx, scriptEnvKey, perms)
}
