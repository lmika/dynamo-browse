package scriptmanager

import (
	"context"
)

type ScriptPlugin struct {
	scriptService      *Service
	name               string
	definedCommands    map[string]*Command
	definedKeyBindings map[string]*Command
	keyToKeyBinding    map[string]string
	relatedItems       []*relatedItemBuilder
}

func (sp *ScriptPlugin) Name() string {
	return sp.name
}

type Command struct {
	plugin *ScriptPlugin
	cmdFn  func(ctx context.Context, args []string) error
}

// Invoke will schedule the command for invocation.  If the script scheduler is free, it will be started immediately.
// Otherwise an error will be returned.
func (c *Command) Invoke(ctx context.Context, args []string, errChan chan error) error {
	return c.plugin.scriptService.sched.runNow(ctx, func(ctx context.Context) {
		errChan <- c.cmdFn(ctx, args)
	})
}

//func (c *Command) LookupRelevantItems(ctx context.Context, table *models.TableInfo, item *models.Item) error {
//
//}
