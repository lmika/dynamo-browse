package ui

import (
	table "github.com/calyptia/go-bubble-table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
	"strings"
)

type uiModel struct {
	table    table.Model
	viewport viewport.Model

	ready     bool
	tableRows []table.Row
	message   string
}

func NewModel() tea.Model {
	tbl := table.New([]string{"seq", "message"}, 100, 20)
	rows := make([]table.Row, 0)
	tbl.SetRows(rows)

	model := uiModel{
		table:     tbl,
		tableRows: rows,
		message:   "",
	}

	return model
}

func (m uiModel) Init() tea.Cmd {
	return nil
}

func (m *uiModel) updateViewportToSelectedMessage() {
	if !m.ready {
		return
	}

	if len(m.tableRows) == 0 {
		return
	}

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

	case tea.WindowSizeMsg:
		footerHeight := lipgloss.Height(m.footerView())

		if !m.ready {
			tableHeight := msg.Height / 2

			m.table.SetSize(msg.Width, tableHeight)
			m.viewport = viewport.New(msg.Width, msg.Height-tableHeight-footerHeight)
			m.viewport.SetContent("(no message selected)")
			m.ready = true
			log.Println("Viewport is now ready")
		} else {
			tableHeight := msg.Height / 2

			m.table.SetSize(msg.Width, tableHeight)
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - tableHeight - footerHeight
			//m.viewport.YPosition = tableHeight
		}

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

	updatedTable, tableMsgs := m.table.Update(msg)
	updatedViewport, viewportMsgs := m.viewport.Update(msg)

	m.table = updatedTable
	m.viewport = updatedViewport

	return m, tea.Batch(tableMsgs, viewportMsgs)
}

func (m uiModel) View() string {
	if !m.ready {
		return "Initializing"
	}

	log.Println("Returning full view")
	return lipgloss.JoinVertical(lipgloss.Top, m.table.View(), m.viewport.View(), m.footerView())
	//return lipgloss.JoinVertical(lipgloss.Top, m.table.View(), m.footerView())
}

func (m uiModel) footerView() string {
	title := m.message
	line := strings.Repeat(" ", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Left, title, line)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
