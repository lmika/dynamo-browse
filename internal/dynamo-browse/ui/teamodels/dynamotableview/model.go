package dynamotableview

import (
	table "github.com/calyptia/go-bubble-table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/dynamoitemview"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
)

var (
	activeHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#4479ff"))
)

type Model struct {
	frameTitle frame.FrameTitle
	table      table.Model
	w, h       int

	// model state
	colOffset int
	rows      []table.Row
	resultSet *models.ResultSet
}

func New() *Model {
	tbl := table.New([]string{"pk", "sk"}, 100, 100)
	rows := make([]table.Row, 0)
	tbl.SetRows(rows)

	frameTitle := frame.NewFrameTitle("No table", true, activeHeaderStyle)

	return &Model{
		frameTitle: frameTitle,
		table:      tbl,
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
		switch msg.String() {
		// Table nav
		case "i", "up":
			m.table.GoUp()
			return m, m.postSelectedItemChanged
		case "k", "down":
			m.table.GoDown()
			return m, m.postSelectedItemChanged
		case "j":
			m.setLeftmostDisplayedColumn(m.colOffset - 1)
			return m, nil
		case "l":
			m.setLeftmostDisplayedColumn(m.colOffset + 1)
			return m, nil
		case "I", "pgup":
			m.table.GoPageUp()
			return m, m.postSelectedItemChanged
		case "K", "pgdn":
			m.table.GoPageDown()
			return m, m.postSelectedItemChanged
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
	m.rebuildTable()
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

	newTbl := table.New(resultSet.Columns[m.colOffset:], m.w, m.h-m.frameTitle.HeaderHeight())
	newRows := make([]table.Row, 0)
	for i, r := range resultSet.Items() {
		if resultSet.Hidden(i) {
			continue
		}

		newRows = append(newRows, itemTableRow{
			resultSet: resultSet,
			itemIndex: i,
			colOffset: m.colOffset,
			item:      r,
		})
	}

	m.rows = newRows
	newTbl.SetRows(newRows)
	for newTbl.Cursor() != m.table.Cursor() {
		if newTbl.Cursor() < m.table.Cursor() {
			newTbl.GoDown()
		} else if newTbl.Cursor() > m.table.Cursor() {
			newTbl.GoUp()
		}
	}
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
