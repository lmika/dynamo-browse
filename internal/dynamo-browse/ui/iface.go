package ui

import tea "github.com/charmbracelet/bubbletea"

type MessagePublisher interface {
	Send(msg tea.Msg)
}
