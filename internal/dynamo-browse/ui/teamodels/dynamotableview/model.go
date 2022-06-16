package dynamotableview

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/dynamoitemview"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	table "github.com/lmika/go-bubble-table"
)

var (
	activeHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#4479ff"))
)

type KeyBinding struct {
	MoveUp   key.Binding
	MoveDown key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding
	ColLeft  key.Binding
	ColRight key.Binding
}

type Model struct {
	frameTitle frame.FrameTitle
	table      table.Model
	w, h       int
	keyBinding KeyBinding

	// model state
	colOffset int
	rows      []table.Row
	resultSet *models.ResultSet
}

type columnModel struct {
	m *Model
}

func (cm columnModel) Len() int {
	return len(cm.m.resultSet.Columns[cm.m.colOffset:])
}

func (cm columnModel) Header(index int) string {
	return cm.m.resultSet.Columns[cm.m.colOffset+index]
}

func New() *Model {
	tbl := table.New(table.SimpleColumns([]string{"pk", "sk"}), 100, 100)
	rows := make([]table.Row, 0)
	tbl.SetRows(rows)

	frameTitle := frame.NewFrameTitle("No table", true, activeHeaderStyle)

	return &Model{
		frameTitle: frameTitle,
		table:      tbl,
		keyBinding: KeyBinding{
			MoveUp:   key.NewBinding(key.WithKeys("i", "up")),
			MoveDown: key.NewBinding(key.WithKeys("k", "down")),
			PageUp:   key.NewBinding(key.WithKeys("I", "pgup")),
			PageDown: key.NewBinding(key.WithKeys("K", "pgdown")),
			Home:     key.NewBinding(key.WithKeys("I", "home")),
			End:      key.NewBinding(key.WithKeys("K", "end")),
			ColLeft:  key.NewBinding(key.WithKeys("j", "left")),
			ColRight: key.NewBinding(key.WithKeys("l", "right")),
		},
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case controllers.NewResultSet:
		m.resultSet = msg.ResultSet
		m.updateTable()
		return m, m.postSelectedItemChanged
	case tea.KeyMsg:
		switch {
		// Table nav
		case key.Matches(msg, m.keyBinding.MoveUp):
			m.table.GoUp()
			return m, m.postSelectedItemChanged
		case key.Matches(msg, m.keyBinding.MoveDown):
			m.table.GoDown()
			return m, m.postSelectedItemChanged
		case key.Matches(msg, m.keyBinding.PageUp):
			m.table.GoPageUp()
			return m, m.postSelectedItemChanged
		case key.Matches(msg, m.keyBinding.PageDown):
			m.table.GoPageDown()
			return m, m.postSelectedItemChanged
		case key.Matches(msg, m.keyBinding.Home):
			m.table.GoTop()
			return m, m.postSelectedItemChanged
		case key.Matches(msg, m.keyBinding.End):
			m.table.GoBottom()
			return m, m.postSelectedItemChanged
		case key.Matches(msg, m.keyBinding.ColLeft):
			m.setLeftmostDisplayedColumn(m.colOffset - 1)
			return m, nil
		case key.Matches(msg, m.keyBinding.ColRight):
			m.setLeftmostDisplayedColumn(m.colOffset + 1)
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) setLeftmostDisplayedColumn(newCol int) {
	if newCol < 0 {
		m.colOffset = 0
	} else if newCol >= len(m.resultSet.Columns) {
		m.colOffset = len(m.resultSet.Columns) - 1
	} else {
		m.colOffset = newCol
	}
	m.table.UpdateView()
}

func (m *Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, m.frameTitle.View(), m.table.View())
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.w, m.h = w, h
	tblHeight := h - m.frameTitle.HeaderHeight()
	m.table.SetSize(w, tblHeight)
	m.frameTitle.Resize(w, h)
	return m
}

func (m *Model) updateTable() {
	m.colOffset = 0

	m.frameTitle.SetTitle("Table: " + m.resultSet.TableInfo.Name)
	m.rebuildTable()
}

func (m *Model) rebuildTable() {
	resultSet := m.resultSet

	newTbl := table.New(columnModel{m}, m.w, m.h-m.frameTitle.HeaderHeight())
	newRows := make([]table.Row, 0)
	for i, r := range resultSet.Items() {
		if resultSet.Hidden(i) {
			continue
		}

		newRows = append(newRows, itemTableRow{
			model:     m,
			resultSet: resultSet,
			itemIndex: i,
			item:      r,
		})
	}

	m.rows = newRows
	newTbl.SetRows(newRows)
	m.table = newTbl
}

func (m *Model) SelectedItemIndex() int {
	selectedItem, ok := m.selectedItem()
	if !ok {
		return -1
	}
	return selectedItem.itemIndex
}

func (m *Model) selectedItem() (itemTableRow, bool) {
	resultSet := m.resultSet
	if resultSet != nil && len(m.rows) > 0 {
		selectedItem, ok := m.table.SelectedRow().(itemTableRow)
		if ok {
			return selectedItem, true
		}
	}

	return itemTableRow{}, false
}

func (m *Model) postSelectedItemChanged() tea.Msg {
	item, ok := m.selectedItem()
	if !ok {
		return nil
	}

	return dynamoitemview.NewItemSelected{ResultSet: item.resultSet, Item: item.item}
}

func (m *Model) Refresh() {

	m.table.SetRows(m.rows)
}
