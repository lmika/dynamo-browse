package keybindings

import "github.com/charmbracelet/bubbles/key"

type KeyBindings struct {
	ColumnPopup *FieldsPopupBinding `keymap:"column-popup"`
	TableView   *TableKeyBinding    `keymap:"item-table"`
	View        *ViewKeyBindings    `keymap:"view"`
}

type FieldsPopupBinding struct {
	Close            key.Binding `keymap:"close"`
	ShiftColumnLeft  key.Binding `keymap:"shift-column-left"`
	ShiftColumnRight key.Binding `keymap:"shift-column-right"`
	ToggleVisible    key.Binding `keymap:"toggle-column-visible"`
	ResetColumns     key.Binding `keymap:"reset-columns"`
	AddColumn        key.Binding `keymap:"add-column"`
	DeleteColumn     key.Binding `keymap:"delete-column"`
}

type TableKeyBinding struct {
	MoveUp   key.Binding `keymap:"move-up"`
	MoveDown key.Binding `keymap:"move-down"`
	PageUp   key.Binding `keymap:"page-up"`
	PageDown key.Binding `keymap:"page-down"`
	Home     key.Binding `keymap:"goto-top"`
	End      key.Binding `keymap:"goto-bottom"`
	ColLeft  key.Binding `keymap:"move-left"`
	ColRight key.Binding `keymap:"move-right"`
}

type ViewKeyBindings struct {
	Mark                 key.Binding `keymap:"mark"`
	CopyItemToClipboard  key.Binding `keymap:"copy-item-to-clipboard"`
	Rescan               key.Binding `keymap:"rescan"`
	PromptForQuery       key.Binding `keymap:"prompt-for-query"`
	PromptForFilter      key.Binding `keymap:"prompt-for-filter"`
	PromptForTable       key.Binding `keymap:"prompt-for-table"`
	ViewBack             key.Binding `keymap:"view-back"`
	ViewForward          key.Binding `keymap:"view-forward"`
	CycleLayoutForward   key.Binding `keymap:"cycle-layout-forward"`
	CycleLayoutBackwards key.Binding `keymap:"cycle-layout-backwards"`
	PromptForCommand     key.Binding `keymap:"prompt-for-command"`
	ShowColumnOverlay    key.Binding `keymap:"show-column-overlay"`
	CancelRunningJob     key.Binding `keymap:"cancel-running-job"`
	Quit                 key.Binding `keymap:"quit"`
}
