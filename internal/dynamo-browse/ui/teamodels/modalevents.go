package teamodels

import tea "github.com/charmbracelet/bubbletea"

// NewModePushed pushes a new mode on the modal stack
type NewModePushed tea.Model

// ModePopped pops a mode from the modal stack
type ModePopped struct{}
