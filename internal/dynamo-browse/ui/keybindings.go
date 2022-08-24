package ui

import "github.com/charmbracelet/bubbles/key"

type KeyBindings struct {
	View *ViewKeyBindings `keymap:"view,group"`
}

type ViewKeyBindings struct {
	Mark                 key.Binding `keymap:"mark"`
	CopyItemToClipboard  key.Binding `keymap:"copy-item-to-clipboard"`
	Rescan               key.Binding `keymap:"rescan"`
	PromptForQuery       key.Binding `keymap:"prompt-for-query"`
	PromptForFilter      key.Binding `keymap:"prompt-for-filter"`
	ViewBack             key.Binding `keymap:"view-back"`
	ViewForward          key.Binding `keymap:"view-forward"`
	CycleLayoutForward   key.Binding `keymap:"cycle-layout-forward"`
	CycleLayoutBackwards key.Binding `keymap:"cycle-layout-backwards"`
	PromptForCommand     key.Binding `keymap:"prompt-for-command"`
	Quit                 key.Binding `keymap:"quit"`
}
