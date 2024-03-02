package scriptmanager

import (
	"context"
	"github.com/risor-io/risor/limits"
)

// scriptEnv is the runtime environment for a particular script execution
type scriptEnv struct {
	filename string
}

type scriptEnvKeyType struct{}

var scriptEnvKey = scriptEnvKeyType{}

func scriptEnvFromCtx(ctx context.Context) scriptEnv {
	perms, _ := ctx.Value(scriptEnvKey).(scriptEnv)
	return perms
}

func ctxWithScriptEnv(ctx context.Context, perms scriptEnv) context.Context {
	newCtx := context.WithValue(ctx, scriptEnvKey, perms)
	newCtx = limits.WithLimits(newCtx, limits.New())
	return newCtx
}
