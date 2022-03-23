package ui

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	table "github.com/calyptia/go-bubble-table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/common/ui/dispatcher"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/common/ui/uimodels"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"strings"
	"text/tabwriter"
)

var (
	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#4479ff"))
)

type uiModel struct {
	table    table.Model
	viewport viewport.Model

	tableWidth, tableHeight int

	ready     bool
	resultSet *models.ResultSet
	message   string

	pendingInput *events.PromptForInput
	textInput    textinput.Model

	dispatcher           *dispatcher.Dispatcher
	tableReadController  *controllers.TableReadController
	tableWriteController *controllers.TableWriteController
}

func NewModel(dispatcher *dispatcher.Dispatcher, tableReadController *controllers.TableReadController, tableWriteController *controllers.TableWriteController) tea.Model {
	tbl := table.New([]string{"pk", "sk"}, 100, 20)
	rows := make([]table.Row, 0)
	tbl.SetRows(rows)

	textInput := textinput.New()

	model := uiModel{
		table:     tbl,
		message:   "Press s to scan",
		textInput: textInput,

		dispatcher:           dispatcher,
		tableReadController:  tableReadController,
		tableWriteController: tableWriteController,
	}

	return model
}

func (m uiModel) Init() tea.Cmd {
	return nil
}

func (m *uiModel) updateTable() {
	if !m.ready {
		return
	}

	newTbl := table.New(m.resultSet.Columns, m.tableWidth, m.tableHeight)
	newRows := make([]table.Row, len(m.resultSet.Items))
	for i, r := range m.resultSet.Items {
		newRows[i] = itemTableRow{m.resultSet, r}
	}
	newTbl.SetRows(newRows)

	m.table = newTbl
}

func (m *uiModel) selectedItem() (itemTableRow, bool) {
	if m.ready && m.resultSet != nil && len(m.resultSet.Items) > 0 {
		selectedItem, ok := m.table.SelectedRow().(itemTableRow)
		if ok {
			return selectedItem, true
		}
	}

	return itemTableRow{}, false
}

func (m *uiModel) updateViewportToSelectedMessage() {
	selectedItem, ok := m.selectedItem()
	if !ok {
		m.viewport.SetContent("(no row selected)")
		return
	}

	viewportContent := &strings.Builder{}
	tabWriter := tabwriter.NewWriter(viewportContent, 0, 1, 1, ' ', 0)
	for _, colName := range selectedItem.resultSet.Columns {
		switch colVal := selectedItem.item[colName].(type) {
		case nil:
			break
		case *types.AttributeValueMemberS:
			fmt.Fprintf(tabWriter, "%v\tS\t%s\n", colName, colVal.Value)
		case *types.AttributeValueMemberN:
			fmt.Fprintf(tabWriter, "%v\tN\t%s\n", colName, colVal.Value)
		default:
			fmt.Fprintf(tabWriter, "%v\t?\t%s\n", colName, "(other)")
		}
	}

	tabWriter.Flush()
	m.viewport.SetContent(viewportContent.String())
}

func (m uiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var textInputCommands tea.Cmd

	switch msg := msg.(type) {

	// Local events
	case controllers.NewResultSet:
		m.resultSet = msg.ResultSet
		m.updateTable()
		m.updateViewportToSelectedMessage()

	// Shared events
	case events.Error:
		m.message = "Error: " + msg.Error()
	case events.Message:
		m.message = string(msg)
	case events.PromptForInput:
		m.textInput.Focus()
		m.textInput.SetValue("")
		m.pendingInput = &msg

	// Tea events
	case tea.WindowSizeMsg:
		fixedViewsHeight := lipgloss.Height(m.headerView()) + lipgloss.Height(m.splitterView()) + lipgloss.Height(m.footerView())
		tableHeight := msg.Height / 2

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-tableHeight-fixedViewsHeight)
			m.viewport.SetContent("(no message selected)")
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - tableHeight - fixedViewsHeight
		}

		m.tableWidth, m.tableHeight = msg.Width, tableHeight
		m.table.SetSize(m.tableWidth, m.tableHeight)

	case tea.KeyMsg:

		// If text input in focus, allow that to accept input messages
		if m.pendingInput != nil {
			switch msg.String() {
			case "ctrl+c", "esc":
				m.pendingInput = nil
			case "enter":
				m.dispatcher.Start(uimodels.WithPromptValue(context.Background(), m.textInput.Value()), m.pendingInput.OnDone)
				m.pendingInput = nil
			default:
				m.textInput, textInputCommands = m.textInput.Update(msg)
			}
			break
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "i":
			m.table.GoUp()
			m.updateViewportToSelectedMessage()
		case "down", "k":
			m.table.GoDown()
			m.updateViewportToSelectedMessage()

		// TODO: these should be moved somewhere else
		case "s":
			m.dispatcher.Start(context.Background(), m.tableReadController.Scan())
		case "D":
			if selectedItem, ok := m.selectedItem(); ok {
				m.dispatcher.Start(context.Background(), m.tableWriteController.Delete(selectedItem.item))
			}
		}
	default:
		m.textInput, textInputCommands = m.textInput.Update(msg)
	}

	updatedTable, tableMsgs := m.table.Update(msg)
	updatedViewport, viewportMsgs := m.viewport.Update(msg)

	m.table = updatedTable
	m.viewport = updatedViewport

	return m, tea.Batch(textInputCommands, tableMsgs, viewportMsgs)
}

func (m uiModel) View() string {
	if !m.ready {
		return "Initializing"
	}

	if m.pendingInput != nil {
		return lipgloss.JoinVertical(lipgloss.Top,
			m.headerView(),
			m.table.View(),
			m.splitterView(),
			m.viewport.View(),
			m.textInput.View(),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		m.headerView(),
		m.table.View(),
		m.splitterView(),
		m.viewport.View(),
		m.footerView(),
	)
}

func (m uiModel) headerView() string {
	title := headerStyle.Render("Table: XXX")
	line := headerStyle.Render(strings.Repeat(" ", max(0, m.viewport.Width-lipgloss.Width(title))))
	return lipgloss.JoinHorizontal(lipgloss.Left, title, line)
}

func (m uiModel) splitterView() string {
	title := headerStyle.Render("Item")
	line := headerStyle.Render(strings.Repeat(" ", max(0, m.viewport.Width-lipgloss.Width(title))))
	return lipgloss.JoinHorizontal(lipgloss.Left, title, line)
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
