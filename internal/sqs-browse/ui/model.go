package ui

import (
	table "github.com/calyptia/go-bubble-table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type uiModel struct {
	table    table.Model
	viewport viewport.Model

	tableRows []table.Row
}

func NewModel() tea.Model {
	tbl := table.New([]string{"seq", "message"}, 100, 20)
	rows := make([]table.Row, 0)
	tbl.SetRows(rows)

	vprt := viewport.New(100, 15)

	model := uiModel{
		table:     tbl,
		viewport:  vprt,
		tableRows: rows,
	}

	return model
}

func (m uiModel) Init() tea.Cmd {
	return nil
}

func (m *uiModel) updateViewportToSelectedMessage() {
	if message, ok := m.table.SelectedRow().(messageTableRow); ok {
		m.viewport.SetContent(message.Data)
	} else {
		m.viewport.SetContent("(no message selected)")
	}
}

func (m uiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NewMessagesEvent:
		for _, newMsg := range msg {
			m.tableRows = append(m.tableRows, messageTableRow(*newMsg))
		}
		m.table.SetRows(m.tableRows)
		m.updateViewportToSelectedMessage()

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "i":
			m.table.GoUp()
			m.updateViewportToSelectedMessage()
		case "down", "k":
			m.table.GoDown()
			m.updateViewportToSelectedMessage()
		}
	}

	updatedTable, tableMsgs := m.table.Update(nil)
	updatedViewport, viewportMsgs := m.viewport.Update(msg)

	m.table = updatedTable
	m.viewport = updatedViewport

	return m, tea.Batch(tableMsgs, viewportMsgs)
}

func (m uiModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, m.table.View(), m.viewport.View())
}
