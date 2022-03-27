package modal

import tea "github.com/charmbracelet/bubbletea"

type newModePushed tea.Model

type modePopped struct{}

// PushMode pushes a new mode on the modal stack.  The new mode will receive keyboard events.
func PushMode(newMode tea.Model) tea.Cmd {
	return func() tea.Msg {
		return newModePushed(newMode)
	}
}

// PopMode pops the top-level mode from the modal stack.  If there's no modes on the stack, this method does nothing.
func PopMode() tea.Msg {
	return modePopped{}
}
