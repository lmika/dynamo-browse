package colselector

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/lmika/audax/internal/dynamo-browse/models/columns"
	"github.com/lmika/audax/internal/dynamo-browse/ui/keybindings"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/layout"
	table "github.com/lmika/go-bubble-table"
)

var style = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63"))

type colListModel struct {
	keyBinding    *keybindings.TableKeyBinding
	colController *controllers.ColumnsController

	rows  []table.Row
	table table.Model
}

func newColListModel(keyBinding *keybindings.TableKeyBinding, colController *controllers.ColumnsController) *colListModel {
	tbl := table.New(table.SimpleColumns([]string{"", "Name"}), 100, 100)
	tbl.SetRows([]table.Row{})

	return &colListModel{
		keyBinding:    keyBinding,
		colController: colController,
		table:         tbl,
	}
}

func (c *colListModel) Init() tea.Cmd {
	return nil
}

func (m *colListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		// Column operations
		case msg.String() == "I":
			return m, events.SetTeaMessage(m.shiftColumnUp(m.table.Cursor()))
		case msg.String() == "K":
			return m, events.SetTeaMessage(m.shiftColumnDown(m.table.Cursor()))
		case msg.String() == " ":
			return m, events.SetTeaMessage(m.colController.ToggleVisible(m.table.Cursor()))
		case msg.String() == "esc":
			return m, events.SetTeaMessage(controllers.HideColumnOverlay{})
		case msg.String() == "R":
			return m, events.SetTeaMessage(m.colController.SetColumnsToResultSet())
		case msg.String() == "a":
			return m, events.SetTeaMessage(m.colController.AddColumn(m.table.Cursor()))
		case msg.String() == "d":
			return m, events.SetTeaMessage(m.colController.DeleteColumn(m.table.Cursor()))

		// Table nav
		case key.Matches(msg, m.keyBinding.MoveUp):
			m.table.GoUp()
			return m, nil
		case key.Matches(msg, m.keyBinding.MoveDown):
			m.table.GoDown()
			return m, nil
		case key.Matches(msg, m.keyBinding.PageUp):
			m.table.GoPageUp()
			return m, nil
		case key.Matches(msg, m.keyBinding.PageDown):
			m.table.GoPageDown()
			return m, nil
		case key.Matches(msg, m.keyBinding.Home):
			m.table.GoTop()
			return m, nil
		case key.Matches(msg, m.keyBinding.End):
			m.table.GoBottom()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (c *colListModel) View() string {
	return style.Width(48).Height(18).Render(c.table.View())
}

func (c *colListModel) Resize(w, h int) layout.ResizingModel {
	c.table.SetSize(46, 16)
	return c
}

func (c *colListModel) refreshTable() {
	colsFromController := c.colController.Columns()
	if len(c.rows) != len(colsFromController.Columns) {
		c.setColumnsFromModel(colsFromController)
	}
	c.table.UpdateView()
}

func (c *colListModel) setColumnsFromModel(cols *columns.Columns) {
	if cols == nil {
		c.table.SetRows([]table.Row{})
		return
	}

	colNames := make([]table.Row, len(cols.Columns))
	for i := range cols.Columns {
		colNames[i] = colListRowModel{c}
	}
	c.rows = colNames
	c.table.SetRows(colNames)

	if c.table.Cursor() >= len(c.rows) {
		c.table.GoBottom()
	}
}

func (c *colListModel) shiftColumnUp(cursor int) tea.Msg {
	msg := c.colController.ShiftColumnLeft(cursor)
	if msg != nil {
		c.table.GoUp()
	}
	return msg
}

func (c *colListModel) shiftColumnDown(cursor int) tea.Msg {
	msg := c.colController.ShiftColumnRight(cursor)
	if msg != nil {
		c.table.GoDown()
	}
	return msg
}
