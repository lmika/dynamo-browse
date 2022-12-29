package scriptmanager

import "context"

type ScriptPlugin struct {
	name            string
	definedCommands map[string]Command
}

func (sp *ScriptPlugin) Name() string {
	return sp.name
}

type Command func(ctx context.Context, args []string) error
