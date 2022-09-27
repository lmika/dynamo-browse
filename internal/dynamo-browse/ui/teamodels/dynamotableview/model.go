package dynamotableview

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/ui/keybindings"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/dynamoitemview"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/styles"
	table "github.com/lmika/go-bubble-table"
	"strings"
)

var (
	activeHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#4479ff"))
)

type Setting interface {
	IsReadOnly() bool
}

type Model struct {
	frameTitle frame.FrameTitle
	table      table.Model
	w, h       int
	keyBinding *keybindings.TableKeyBinding
	setting    Setting

	// model state
	isReadOnly bool
	colOffset  int
	rows       []table.Row
	resultSet  *models.ResultSet
}

func New(keyBinding *keybindings.TableKeyBinding, setting Setting, uiStyles styles.Styles) *Model {
	tbl := table.New(table.SimpleColumns([]string{"pk", "sk"}), 100, 100)
	rows := make([]table.Row, 0)
	tbl.SetRows(rows)

	frameTitle := frame.NewFrameTitle("No table", true, uiStyles.Frames)
	isReadOnly := setting.IsReadOnly()

	return &Model{
		isReadOnly: isReadOnly,
		frameTitle: frameTitle,
		table:      tbl,
		keyBinding: keyBinding,
		setting:    setting,
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
	} else if newCol >= len(m.resultSet.Columns()) {
		m.colOffset = len(m.resultSet.Columns()) - 1
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

	tableName := new(strings.Builder)
	tableName.WriteString("Table: " + m.resultSet.TableInfo.Name)
	if m.setting.IsReadOnly() {
		tableName.WriteString(" [RO]")
	}

	m.frameTitle.SetTitle(tableName.String())
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
		return dynamoitemview.NewItemSelected{ResultSet: item.resultSet, Item: nil}
	}

	return dynamoitemview.NewItemSelected{ResultSet: item.resultSet, Item: item.item}
}

func (m *Model) Refresh() tea.Cmd {
	m.table.SetRows(m.rows)
	return m.postSelectedItemChanged
}
