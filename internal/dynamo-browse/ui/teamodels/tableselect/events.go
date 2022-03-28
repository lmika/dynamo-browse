package tableselect

import tea "github.com/charmbracelet/bubbletea"

func IndicateLoadingTables() tea.Cmd {
	return func() tea.Msg {
		return indicateLoadingTablesMsg{}
	}
}
func ShowTableSelect(onSelected func(n string) tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		return showTableSelectMsg{
			onSelected: onSelected,
		}
	}
}

type indicateLoadingTablesMsg struct{}

type showTableSelectMsg struct {
	onSelected func(n string) tea.Cmd
}
