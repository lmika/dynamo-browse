package tableselect

import tea "github.com/charmbracelet/bubbletea"

type indicateLoadingTablesMsg struct{}

type showTableSelectMsg struct {
	onSelected func(n string) tea.Cmd
}
