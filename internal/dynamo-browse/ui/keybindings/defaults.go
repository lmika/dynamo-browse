package keybindings

import "github.com/charmbracelet/bubbles/key"

func Default() *KeyBindings {
	return &KeyBindings{
		ColumnPopup: &FieldsPopupBinding{
			Close:            key.NewBinding(key.WithKeys("ctrl+c", "esc"), key.WithHelp("ctrl+c/esc", "close popup")),
			ShiftColumnLeft:  key.NewBinding(key.WithKeys("I", "shift column left")),
			ShiftColumnRight: key.NewBinding(key.WithKeys("K", "shift column right")),
			ToggleVisible:    key.NewBinding(key.WithKeys(" ", "toggle column visible")),
			ResetColumns:     key.NewBinding(key.WithKeys("R", "reset columns")),
			AddColumn:        key.NewBinding(key.WithKeys("a", "add new column")),
			DeleteColumn:     key.NewBinding(key.WithKeys("d", "delete column")),
		},
		TableView: &TableKeyBinding{
			MoveUp:   key.NewBinding(key.WithKeys("i", "up")),
			MoveDown: key.NewBinding(key.WithKeys("k", "down")),
			PageUp:   key.NewBinding(key.WithKeys("I", "pgup")),
			PageDown: key.NewBinding(key.WithKeys("K", "pgdown")),
			Home:     key.NewBinding(key.WithKeys("0", "home")),
			End:      key.NewBinding(key.WithKeys("$", "end")),
			ColLeft:  key.NewBinding(key.WithKeys("j", "left")),
			ColRight: key.NewBinding(key.WithKeys("l", "right")),
		},
		View: &ViewKeyBindings{
			Mark:                 key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "mark")),
			ToggleMarkedItems:    key.NewBinding(key.WithKeys("M"), key.WithHelp("M", "toggle marged items")),
			CopyItemToClipboard:  key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy item to clipboard")),
			CopyTableToClipboard: key.NewBinding(key.WithKeys("C"), key.WithHelp("C", "copy table to clipboard")),
			Rescan:               key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "rescan")),
			PromptForQuery:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "prompt for query")),
			PromptForFilter:      key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
			FetchNextPage:        key.NewBinding(key.WithKeys(">"), key.WithHelp(">", "fetch next page")),
			ViewBack:             key.NewBinding(key.WithKeys("backspace"), key.WithHelp("backspace", "go back")),
			ViewForward:          key.NewBinding(key.WithKeys("\\"), key.WithHelp("\\", "go forward")),
			CycleLayoutForward:   key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "cycle layout forward")),
			CycleLayoutBackwards: key.NewBinding(key.WithKeys("W"), key.WithHelp("W", "cycle layout backward")),
			PromptForCommand:     key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "prompt for command")),
			ShowColumnOverlay:    key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "show column overlay")),
			ShowRelItemsOverlay:  key.NewBinding(key.WithKeys("O"), key.WithHelp("O", "show related items overlay")),
			CancelRunningJob:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "cancel running job or quit")),
			Quit:                 key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
		},
	}
}
