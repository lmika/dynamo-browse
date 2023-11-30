package relselector

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/dynamo-browse/internal/common/sliceutils"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/controllers"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/relitems"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/utils"
)

var (
	frameColor = lipgloss.Color("63")

	frameStyle = lipgloss.NewStyle().
			Foreground(frameColor)
	style = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(frameColor)

	keyEnter = key.NewBinding(key.WithKeys(tea.KeyEnter.String()))
)

type listModel struct {
	list   list.Model
	height int
}

func newListModel() *listModel {
	items := []list.Item{}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#2c5fb7")).
		Foreground(lipgloss.Color("#2c5fb7")).
		Padding(0, 0, 0, 1)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#2c5fb7")).
		Foreground(lipgloss.Color("#5277b7")).
		Padding(0, 0, 0, 1)

	list := list.New(items, delegate, overlayWidth, overlayHeight-4)
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
	list.SetShowHelp(false)
	//list.DisableQuitKeybindings()

	return &listModel{
		list: list,
	}
}

func (m *listModel) Init() tea.Cmd {
	return nil
}

func (m *listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cc utils.CmdCollector

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keyEnter):
			return m, events.SetTeaMessage(controllers.HideColumnOverlay{})
		default:
			m.list = cc.Collect(m.list.Update(msg)).(list.Model)
		}
	default:
		m.list = cc.Collect(m.list.Update(msg)).(list.Model)
	}
	return m, cc.Cmd()
}

func (m *listModel) View() string {
	innerView := lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.PlaceHorizontal(overlayWidth-2, lipgloss.Center, "Related Items"),
		frameStyle.Render(strings.Repeat(lipgloss.NormalBorder().Top, overlayWidth-2)),
		m.list.View(),
	)

	view := style.Width(overlayWidth - 2).Height(m.height - 2).Render(innerView)

	return view
}

func (m *listModel) Resize(w, h int) layout.ResizingModel {
	return m
}

func (m *listModel) setItems(items []relitems.RelatedItem, newHeight int) {
	listItems := sliceutils.Map(items, func(item relitems.RelatedItem) list.Item {
		return relItemModel{name: item.Name}
	})
	m.list.SetItems(listItems)
	m.list.Select(0)
	m.list.SetHeight(newHeight - 4)

	m.height = newHeight
}
