package uimodels

import tea "github.com/charmbracelet/bubbletea"

type UIContext interface {
	Send(teaMessage tea.Msg)
	Message(msg string)
	Messagef(format string, args ...interface{})
	Input(prompt string, onDone Operation)
}
