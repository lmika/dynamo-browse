package colselector

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/controllers"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/columns"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/keybindings"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/ui/teamodels/layout"
	table "github.com/lmika/go-bubble-table"
	"strings"
)

var frameColor = lipgloss.Color("63")

var frameStyle = lipgloss.NewStyle().
	Foreground(frameColor)
var style = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(frameColor)

type colListModel struct {
	keyBinding    *keybindings.KeyBindings
	colController *controllers.ColumnsController

	rows  []table.Row
	table table.Model
}

func newColListModel(keyBinding *keybindings.KeyBindings, colController *controllers.ColumnsController) *colListModel {
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
	case controllers.SetSelectedColumnInColSelector:
		// HACK: this needs to work for all cases
		if int(msg) == m.table.Cursor()+1 {
			m.table.GoDown()
		}
	case tea.KeyMsg:
		switch {
		// Column operations
		case key.Matches(msg, m.keyBinding.ColumnPopup.ShiftColumnLeft):
			return m, events.SetTeaMessage(m.shiftColumnUp(m.table.Cursor()))
		case key.Matches(msg, m.keyBinding.ColumnPopup.ShiftColumnRight):
			return m, events.SetTeaMessage(m.shiftColumnDown(m.table.Cursor()))
		case key.Matches(msg, m.keyBinding.ColumnPopup.ToggleVisible):
			return m, events.SetTeaMessage(m.colController.ToggleVisible(m.table.Cursor()))
		case key.Matches(msg, m.keyBinding.ColumnPopup.Close):
			return m, events.SetTeaMessage(controllers.HideColumnOverlay{})
		case key.Matches(msg, m.keyBinding.ColumnPopup.ResetColumns):
			return m, events.SetTeaMessage(m.colController.SetColumnsToResultSet())
		case key.Matches(msg, m.keyBinding.ColumnPopup.AddColumn):
			return m, events.SetTeaMessage(m.colController.AddColumn(m.table.Cursor()))
		case key.Matches(msg, m.keyBinding.ColumnPopup.DeleteColumn):
			return m, events.SetTeaMessage(m.colController.DeleteColumn(m.table.Cursor()))

		// Main table nav
		case key.Matches(msg, m.keyBinding.TableView.ColLeft):
			return m, events.SetTeaMessage(controllers.MoveLeftmostDisplayedColumnInTableViewBy(-1))
		case key.Matches(msg, m.keyBinding.TableView.ColRight):
			return m, events.SetTeaMessage(controllers.MoveLeftmostDisplayedColumnInTableViewBy(1))

		// Table nav
		case key.Matches(msg, m.keyBinding.TableView.MoveUp):
			m.table.GoUp()
			return m, nil
		case key.Matches(msg, m.keyBinding.TableView.MoveDown):
			m.table.GoDown()
			return m, nil
		case key.Matches(msg, m.keyBinding.TableView.PageUp):
			m.table.GoPageUp()
			return m, nil
		case key.Matches(msg, m.keyBinding.TableView.PageDown):
			m.table.GoPageDown()
			return m, nil
		case key.Matches(msg, m.keyBinding.TableView.Home):
			m.table.GoTop()
			return m, nil
		case key.Matches(msg, m.keyBinding.TableView.End):
			m.table.GoBottom()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (c *colListModel) View() string {
	innerView := lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.PlaceHorizontal(overlayWidth-2, lipgloss.Center, "Columns"),
		frameStyle.Render(strings.Repeat(lipgloss.NormalBorder().Top, 48)),
		c.table.View(),
	)

	view := style.Width(overlayWidth - 2).Height(overlayHeight - 2).Render(innerView)

	return view
}

func (c *colListModel) Resize(w, h int) layout.ResizingModel {
	c.table.SetSize(overlayWidth-4, overlayHeight-4)
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
