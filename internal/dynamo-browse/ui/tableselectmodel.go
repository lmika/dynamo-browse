package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type tableSelectModel struct {
	list list.Model
}

func (t tableSelectModel) Init() tea.Cmd {
	return nil
}

func (t tableSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.list.SetHeight(msg.Height)
		t.list.SetWidth(msg.Width)
		return t, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return t, tea.Quit

		case "enter":
			//i, ok := m.list.SelectedItem().(item)
			//if ok {
			//	m.choice = string(i)
			//}
			return t, tea.Quit
		}
	}

	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return t, cmd
}

func (t tableSelectModel) View() string {
	return t.list.View()
}

func newTableSelectModel(w, h int) tableSelectModel {
	tableItems := []tableItem{
		{name: "alpha"},
		{name: "beta"},
		{name: "gamma"},
	}

	items := toListItems(tableItems)

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	return tableSelectModel{
		list: list.New(items, delegate, w, h),
	}
}

type tableItem struct {
	name string
}

func (ti tableItem) FilterValue() string {
	return ""
}

func (ti tableItem) Title() string {
	return ti.name
}

func (ti tableItem) Description() string {
	return "abc"
}

func toListItems[T list.Item](xs []T) []list.Item {
	ls := make([]list.Item, len(xs))
	for i, x := range xs {
		ls[i] = x
	}
	return ls
}
