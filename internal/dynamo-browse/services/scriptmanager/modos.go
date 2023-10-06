package scriptmanager

import (
	"context"
	"github.com/risor-io/risor/object"
	"os"
	"os/exec"
)

type osModule struct {
}

func (om *osModule) exec(ctx context.Context, args ...object.Object) object.Object {
	if err := require("os.exec", 1, args); err != nil {
		return err
	}

	cmdExec, objErr := object.AsString(args[0])
	if objErr != nil {
		return objErr
	}

	opts := scriptEnvFromCtx(ctx).options
	if !opts.Permissions.AllowShellCommands {
		return object.Errorf("permission error: no permission to shell out")
	}

	cmd := exec.Command(opts.OSExecShell, "-c", cmdExec)
	out, err := cmd.Output()
	if err != nil {
		return object.NewError(err)
	}

	return object.NewString(string(out))
}

func (om *osModule) env(ctx context.Context, args ...object.Object) object.Object {
	if err := require("os.env", 1, args); err != nil {
		return err
	}

	cmdEnvName, objErr := object.AsString(args[0])
	if objErr != nil {
		return objErr
	}

	opts := scriptEnvFromCtx(ctx).options
	if !opts.Permissions.AllowEnv {
		return object.Nil
	}

	envVal, hasVal := os.LookupEnv(cmdEnvName)
	if !hasVal {
		return object.Nil
	}
	return object.NewString(envVal)
}

func (om *osModule) register() *object.Module {
	return object.NewBuiltinsModule("os", map[string]object.Object{
		"exec": object.NewBuiltin("exec", om.exec),
		"env":  object.NewBuiltin("env", om.env),
	})
}
