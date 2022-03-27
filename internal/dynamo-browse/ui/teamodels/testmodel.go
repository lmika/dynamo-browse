package teamodels

import (
	tea "github.com/charmbracelet/bubbletea"
)

// TestModel is a model used for testing
type TestModel struct {
	Message      string
	OnKeyPressed func(k string) tea.Cmd
}

func (t TestModel) Init() tea.Cmd {
	return nil
}

func (t TestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return t, tea.Quit
		default:
			return t, t.OnKeyPressed(msg.String())
		}
	}

	return t, nil
}

func (t TestModel) View() string {
	return t.Message
}
