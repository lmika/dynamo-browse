package scriptmanager

import (
	"context"
	"github.com/cloudcmds/tamarin/arg"
	"github.com/cloudcmds/tamarin/object"
	"github.com/cloudcmds/tamarin/scope"
	"os"
	"os/exec"
)

type osModule struct {
}

func (om *osModule) exec(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("os.exec", 1, args); err != nil {
		return err
	}

	cmdExec, objErr := object.AsString(args[0])
	if objErr != nil {
		return objErr
	}

	opts := optionFromCtx(ctx)
	if !opts.Permissions.AllowShellCommands {
		return object.NewErrResult(object.Errorf("permission error: no permission to shell out"))
	}

	cmd := exec.Command(opts.OSExecShell, "-c", cmdExec)
	out, err := cmd.Output()
	if err != nil {
		return object.NewErrResult(object.NewError(err))
	}

	return object.NewOkResult(object.NewString(string(out)))
}

func (om *osModule) env(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("os.env", 1, args); err != nil {
		return err
	}

	cmdEnvName, objErr := object.AsString(args[0])
	if objErr != nil {
		return objErr
	}

	opts := optionFromCtx(ctx)
	if !opts.Permissions.AllowEnv {
		return object.Nil
	}

	envVal, hasVal := os.LookupEnv(cmdEnvName)
	if !hasVal {
		return object.Nil
	}
	return object.NewString(envVal)
}

func (om *osModule) register(scp *scope.Scope) {
	modScope := scope.New(scope.Opts{})
	mod := object.NewModule("os", modScope)

	modScope.AddBuiltins([]*object.Builtin{
		object.NewBuiltin("exec", om.exec, mod),
		object.NewBuiltin("env", om.env, mod),
	})

	scp.Declare("os", mod, true)
}
