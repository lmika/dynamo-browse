package ui

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	table "github.com/calyptia/go-bubble-table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"log"
	"strings"
	"text/tabwriter"
)

type uiModel struct {
	table    table.Model
	viewport viewport.Model

	msgPublisher MessagePublisher
	tableService *tables.Service
	tableName    string

	tableWidth, tableHeight int

	ready     bool
	resultSet *models.ResultSet
	message   string
}

func NewModel(tableService *tables.Service, msgPublisher MessagePublisher, tableName string) tea.Model {
	tbl := table.New([]string{"pk", "sk"}, 100, 20)
	rows := make([]table.Row, 0)
	tbl.SetRows(rows)

	model := uiModel{
		table:        tbl,
		tableService: tableService,
		tableName:    tableName,
		msgPublisher: msgPublisher,
		message:      "Press s to scan",
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

func (m *uiModel) updateViewportToSelectedMessage() {
	if !m.ready {
		return
	}

	if m.resultSet == nil || len(m.resultSet.Items) == 0 {
		return
	}

	selectedItem, ok := m.table.SelectedRow().(itemTableRow)
	if !ok {
		m.viewport.SetContent("(no row selected)")
		return
	}

	viewportContent := &strings.Builder{}
	tabWriter := tabwriter.NewWriter(viewportContent, 0, 1, 1, ' ', 0)
	for _, colName := range selectedItem.resultSet.Columns {
		fmt.Fprintf(tabWriter, "%v\t", colName)

		switch colVal := selectedItem.item[colName].(type) {
		case nil:
			fmt.Fprintln(tabWriter, "(nil)")
		case *types.AttributeValueMemberS:
			fmt.Fprintln(tabWriter, colVal.Value)
		case *types.AttributeValueMemberN:
			fmt.Fprintln(tabWriter, colVal.Value)
		default:
			fmt.Fprintln(tabWriter, "(other)")
		}
	}

	tabWriter.Flush()
	m.viewport.SetContent(viewportContent.String())
}

func (m uiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case setStatusMessage:
		m.message = ""
	case errorRaised:
		m.message = "Error: " + msg.Error()
	case newResultSet:
		m.resultSet = msg.ResultSet
		m.updateTable()
		m.updateViewportToSelectedMessage()
	case tea.WindowSizeMsg:
		footerHeight := lipgloss.Height(m.footerView())
		tableHeight := msg.Height / 2

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-tableHeight-footerHeight)
			m.viewport.SetContent("(no message selected)")
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - tableHeight - footerHeight
		}

		m.tableWidth, m.tableHeight = msg.Width, tableHeight
		m.table.SetSize(m.tableWidth, m.tableHeight)

	case tea.KeyMsg:

		switch msg.String() {
		case "s":
			m.startOperation("Scanning...", func(ctx context.Context) (tea.Msg, error) {
				resultSet, err := m.tableService.Scan(ctx, m.tableName)
				if err != nil {
					return nil, err
				}
				return newResultSet{resultSet}, nil
			})
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

// TODO: this should probably be a separate service
func (m *uiModel) startOperation(msg string, op func(ctx context.Context) (tea.Msg, error)) {
	m.message = msg
	go func() {
		resMsg, err := op(context.Background())
		if err != nil {
			m.msgPublisher.Send(errorRaised(err))
		} else if resMsg != nil {
			m.msgPublisher.Send(resMsg)
		}
		m.msgPublisher.Send(setStatusMessage(""))
	}()
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
