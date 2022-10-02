package colselector

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/ui/keybindings"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/layout"
	table "github.com/lmika/go-bubble-table"
)

var style = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63"))

type colListModel struct {
	keyBinding *keybindings.TableKeyBinding

	table table.Model
}

func newColListModel(keyBinding *keybindings.TableKeyBinding) *colListModel {
	tbl := table.New(table.SimpleColumns([]string{"", "Name"}), 100, 100)
	tbl.SetRows([]table.Row{})

	return &colListModel{
		keyBinding: keyBinding,
		table:      tbl,
	}
}

func (c *colListModel) Init() tea.Cmd {
	return nil
}

func (m *colListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
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

func (c *colListModel) setColumnsFromModel(cols *models.Columns) {
	colNames := make([]table.Row, len(cols.Columns))
	for i := range cols.Columns {
		colNames[i] = table.SimpleRow{".", cols.Columns[i]}
	}
	c.table.SetRows(colNames)
}
