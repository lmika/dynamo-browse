package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/commandctrl"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/dynamoitemedit"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/dynamoitemview"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/dynamotableview"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/statusandprompt"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/tableselect"
)

type Model struct {
	tableReadController  *controllers.TableReadController
	tableWriteController *controllers.TableWriteController
	commandController    *commandctrl.CommandController
	itemEdit             *dynamoitemedit.Model
	statusAndPrompt      *statusandprompt.StatusAndPrompt
	tableSelect          *tableselect.Model

	root      tea.Model
	tableView *dynamotableview.Model
}

func NewModel(rc *controllers.TableReadController, wc *controllers.TableWriteController, cc *commandctrl.CommandController) Model {
	dtv := dynamotableview.New()
	div := dynamoitemview.New()
	mainView := layout.NewVBox(layout.LastChildFixedAt(17), dtv, div)

	itemEdit := dynamoitemedit.NewModel(mainView)
	statusAndPrompt := statusandprompt.New(itemEdit, "")
	tableSelect := tableselect.New(statusAndPrompt)

	cc.AddCommands(&commandctrl.CommandContext{
		Commands: map[string]commandctrl.Command{
			"q": commandctrl.NoArgCommand(tea.Quit),
			"table": func(args []string) tea.Cmd {
				if len(args) == 0 {
					return rc.ListTables()
				} else {
					return rc.ScanTable(args[0])
				}
			},
			"unmark": commandctrl.NoArgCommand(rc.Unmark()),
			"delete": commandctrl.NoArgCommand(wc.DeleteMarked()),
		},
	})

	root := layout.FullScreen(tableSelect)

	return Model{
		tableReadController:  rc,
		tableWriteController: wc,
		commandController:    cc,
		itemEdit:             itemEdit,
		statusAndPrompt:      statusAndPrompt,
		tableSelect:          tableSelect,
		root:                 root,
		tableView:            dtv,
	}
}

func (m Model) Init() tea.Cmd {
	return m.tableReadController.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case controllers.ResultSetUpdated:
		m.tableView.Refresh()
	case tea.KeyMsg:
		if !m.statusAndPrompt.InPrompt() && !m.tableSelect.Visible() {
			switch msg.String() {
			case "m":
				if idx := m.tableView.SelectedItemIndex(); idx >= 0 {
					return m, m.tableWriteController.ToggleMark(idx)
				}
			case "s":
				return m, m.tableReadController.Rescan()
			case "/":
				return m, m.tableReadController.Filter()
			case "e":
				m.itemEdit.Visible()
				return m, nil
			case ":":
				return m, m.commandController.Prompt()
			case "ctrl+c", "esc":
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.root, cmd = m.root.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.root.View()
}
