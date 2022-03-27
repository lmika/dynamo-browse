package dynamotableview

import (
	table "github.com/calyptia/go-bubble-table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/dynamoitemview"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
)

type Model struct {
	tableReadControllers *controllers.TableReadController
	table                table.Model
	w, h                 int

	// model state
	resultSet *models.ResultSet
}

func New(tableReadControllers *controllers.TableReadController) Model {
	tbl := table.New([]string{"pk", "sk"}, 100, 100)
	rows := make([]table.Row, 0)
	tbl.SetRows(rows)

	return Model{tableReadControllers: tableReadControllers, table: tbl}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		// TEMP
		case "s":
			return m, m.tableReadControllers.Scan()
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string {
	return m.table.View()
}

func (m Model) Resize(w, h int) layout.ResizingModel {
	m.w, m.h = w, h
	m.table.SetSize(w, h)
	return m
}

func (m *Model) updateTable() {
	resultSet := m.resultSet

	newTbl := table.New(resultSet.Columns, m.w, m.h)
	newRows := make([]table.Row, len(resultSet.Items))
	for i, r := range resultSet.Items {
		newRows[i] = itemTableRow{resultSet, r}
	}
	newTbl.SetRows(newRows)

	m.table = newTbl
}

func (m *Model) selectedItem() (itemTableRow, bool) {
	resultSet := m.resultSet
	if resultSet != nil && len(resultSet.Items) > 0 {
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

/*
func (m *Model) updateViewportToSelectedMessage() {
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
*/
