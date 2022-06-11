package tableselect

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type listController struct {
	list list.Model
}

func newListController(tableNames []string, w, h int) listController {
	items := toListItems(tableNames)

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#2c5fb7")).
		Foreground(lipgloss.Color("#2c5fb7")).
		Padding(0, 0, 0, 1)

	list := list.New(items, delegate, w, h)
	list.KeyMap.CursorUp = key.NewBinding(
		key.WithKeys("up", "i"),
		key.WithHelp("↑/i", "up"),
	)
	list.KeyMap.CursorDown = key.NewBinding(
		key.WithKeys("down", "k"),
		key.WithHelp("↓/k", "down"),
	)
	list.KeyMap.PrevPage = key.NewBinding(
		key.WithKeys("left", "j", "pgup", "b", "u"),
		key.WithHelp("←/j/pgup", "prev page"),
	)
	list.KeyMap.NextPage = key.NewBinding(
		key.WithKeys("right", "l", "pgdown", "f", "d"),
		key.WithHelp("→/l/pgdn", "next page"),
	)
	list.SetShowTitle(false)

	return listController{list: list}
}

func (l listController) Init() tea.Cmd {
	return nil
}

func (l listController) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newList, cmd := l.list.Update(msg)
	l.list = newList
	return l, cmd
}

func (l listController) View() string {
	return l.list.View()
}

func (l listController) Resize(w, h int) layout.ResizingModel {
	l.list.SetSize(w, h)
	return l
}
