package utils

import tea "github.com/charmbracelet/bubbletea"

type Updatable[T any] interface {
	Update(msg tea.Msg) (T, tea.Cmd)
}

func Update[T Updatable[T]](model T, msg tea.Msg) (T, tea.Cmd) {
	newModel, cmd := model.Update(msg)
	return newModel, cmd
}
