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
	"github.com/lmika/awstools/internal/common/ui/commandctrl"
	"github.com/lmika/awstools/internal/common/ui/dispatcher"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/common/ui/uimodels"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"strings"
	"text/tabwriter"
)

var (
	activeHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#4479ff"))

	inactiveHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#d1d1d1"))
)

type uiModel struct {
	table    table.Model
	viewport viewport.Model

	tableWidth, tableHeight int

	ready bool
	//resultSet *models.ResultSet
	state   controllers.State
	message string

	pendingInput *events.PromptForInput
	textInput    textinput.Model

	dispatcher           *dispatcher.Dispatcher
	commandController    *commandctrl.CommandController
	tableReadController  *controllers.TableReadController
	tableWriteController *controllers.TableWriteController
}

func NewModel(dispatcher *dispatcher.Dispatcher, commandController *commandctrl.CommandController, tableReadController *controllers.TableReadController, tableWriteController *controllers.TableWriteController) tea.Model {
	tbl := table.New([]string{"pk", "sk"}, 100, 20)
	rows := make([]table.Row, 0)
	tbl.SetRows(rows)

	textInput := textinput.New()

	model := uiModel{
		table:     tbl,
		message:   "Press s to scan",
		textInput: textInput,

		dispatcher:           dispatcher,
		commandController:    commandController,
		tableReadController:  tableReadController,
		tableWriteController: tableWriteController,
	}

	return model
}

func (m uiModel) Init() tea.Cmd {
	m.invokeOperation(context.Background(), m.tableReadController.Scan())
	return nil
}

func (m *uiModel) updateTable() {
	if !m.ready {
		return
	}

	resultSet := m.state.ResultSet
	newTbl := table.New(resultSet.Columns, m.tableWidth, m.tableHeight)
	newRows := make([]table.Row, len(resultSet.Items))
	for i, r := range resultSet.Items {
		newRows[i] = itemTableRow{resultSet, r}
	}
	newTbl.SetRows(newRows)

	m.table = newTbl
}

func (m *uiModel) selectedItem() (itemTableRow, bool) {
	resultSet := m.state.ResultSet
	if m.ready && resultSet != nil && len(resultSet.Items) > 0 {
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
		m.state.ResultSet = msg.ResultSet
		m.updateTable()
		m.updateViewportToSelectedMessage()
	case controllers.SetReadWrite:
		m.state.InReadWriteMode = msg.NewValue

	// Shared events
	case events.Error:
		m.message = "Error: " + msg.Error()
	case events.Message:
		m.message = string(msg)
	case events.PromptForInput:
		m.textInput.Prompt = msg.Prompt
		m.textInput.Focus()
		m.textInput.SetValue("")
		m.pendingInput = &msg

	// Tea events
	case tea.WindowSizeMsg:
		fixedViewsHeight := lipgloss.Height(m.headerView()) + lipgloss.Height(m.splitterView()) + lipgloss.Height(m.footerView())
		viewportHeight := msg.Height / 2		// TODO: make this dynamic
		if viewportHeight > 15 {
			viewportHeight = 15
		}
		tableHeight := msg.Height - fixedViewsHeight - viewportHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewportHeight)
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
				m.invokeOperation(uimodels.WithPromptValue(context.Background(), m.textInput.Value()), m.pendingInput.OnDone)
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
		case ":":
			m.invokeOperation(context.Background(), m.commandController.Prompt())
		case "s":
			m.invokeOperation(context.Background(), m.tableReadController.Scan())
		case "D":
			m.invokeOperation(context.Background(), m.tableWriteController.Delete())
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

func (m uiModel) invokeOperation(ctx context.Context, op uimodels.Operation) {
	state := m.state
	if selectedItem, ok := m.selectedItem(); ok {
		state.SelectedItem = selectedItem.item
	}

	ctx = controllers.ContextWithState(ctx, state)
	m.dispatcher.Start(ctx, op)
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
	var titleText string
	if m.state.ResultSet != nil {
		titleText = "Table: " + m.state.ResultSet.TableInfo.Name
	} else {
		titleText = "No table"
	}

	title := activeHeaderStyle.Render(titleText)
	line := activeHeaderStyle.Render(strings.Repeat(" ", max(0, m.viewport.Width-lipgloss.Width(title))))
	return lipgloss.JoinHorizontal(lipgloss.Left, title, line)
}

func (m uiModel) splitterView() string {
	title := inactiveHeaderStyle.Render("Item")
	line := inactiveHeaderStyle.Render(strings.Repeat(" ", max(0, m.viewport.Width-lipgloss.Width(title))))
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
