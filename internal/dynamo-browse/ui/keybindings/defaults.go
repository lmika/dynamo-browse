package keybindings

import "github.com/charmbracelet/bubbles/key"

func Default() *KeyBindings {
	return &KeyBindings{
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
			CopyItemToClipboard:  key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy item to clipboard")),
			Rescan:               key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "rescan")),
			PromptForQuery:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "prompt for query")),
			PromptForFilter:      key.NewBinding(key.WithKeys("f"), key.WithHelp("/", "filter")),
			ViewBack:             key.NewBinding(key.WithKeys("backspace"), key.WithHelp("backspace", "go back")),
			ViewForward:          key.NewBinding(key.WithKeys("\\"), key.WithHelp("\\", "go forward")),
			CycleLayoutForward:   key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "cycle layout forward")),
			CycleLayoutBackwards: key.NewBinding(key.WithKeys("W"), key.WithHelp("W", "cycle layout backward")),
			PromptForCommand:     key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "prompt for command")),
			Quit:                 key.NewBinding(key.WithKeys("ctrl+c", "esc"), key.WithHelp("ctrl+c/esc", "quit")),
		},
	}
}
