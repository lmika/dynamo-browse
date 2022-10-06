package dynamotableview

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/columns"
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

type ColumnsProvider interface {
	Columns() *columns.Columns
}

type Model struct {
	frameTitle      frame.FrameTitle
	table           table.Model
	w, h            int
	keyBinding      *keybindings.TableKeyBinding
	setting         Setting
	columnsProvider ColumnsProvider

	// model state
	isReadOnly bool
	colOffset  int
	rows       []table.Row
	columns    []columns.Column
	resultSet  *models.ResultSet
}

func New(keyBinding *keybindings.TableKeyBinding, columnsProvider ColumnsProvider, setting Setting, uiStyles styles.Styles) *Model {
	frameTitle := frame.NewFrameTitle("No table", true, uiStyles.Frames)
	isReadOnly := setting.IsReadOnly()

	model := &Model{
		isReadOnly:      isReadOnly,
		frameTitle:      frameTitle,
		keyBinding:      keyBinding,
		setting:         setting,
		columnsProvider: columnsProvider,
	}

	model.table = table.New(columnModel{model}, 100, 100)
	model.table.SetRows([]table.Row{})

	return model
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
	case controllers.ColumnsUpdated:
		m.rebuildTable(&m.table)
		return m, m.postSelectedItemChanged
	case controllers.SettingsUpdated:
		m.updateTableHeading()
		return m, nil
	case controllers.MoveLeftmostDisplayedColumnInTableViewBy:
		m.setLeftmostDisplayedColumn(m.colOffset + int(msg))
		return m, nil
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
	} else if newCol >= len(m.columnsProvider.Columns().Columns) {
		m.colOffset = len(m.columnsProvider.Columns().Columns) - 1
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

func (m *Model) updateTableHeading() {
	tableName := new(strings.Builder)
	tableName.WriteString("Table: " + m.resultSet.TableInfo.Name)
	if m.setting.IsReadOnly() {
		tableName.WriteString(" [RO]")
	}

	m.frameTitle.SetTitle(tableName.String())
}

func (m *Model) updateTable() {
	m.updateTableHeading()
	m.colOffset = 0
	m.rebuildTable(nil)
}

func (m *Model) rebuildTable(targetTbl *table.Model) {
	var tbl table.Model

	resultSet := m.resultSet

	// Use the target table model if you can, but if it's nil or the number of rows is smaller than the
	// existing table, create a new one
	if targetTbl == nil || len(resultSet.Items()) > len(m.rows) {
		tbl = table.New(columnModel{m}, m.w, m.h-m.frameTitle.HeaderHeight())
		if targetTbl != nil {
			tbl.GoBottom()
		}
	} else {
		tbl = *targetTbl
	}

	m.columns = m.columnsProvider.Columns().VisibleColumns()

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
	tbl.SetRows(newRows)

	m.table = tbl
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

	// Fix bug??
	if m.table.Cursor() < 0 {
		return itemTableRow{}, false
	}

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
