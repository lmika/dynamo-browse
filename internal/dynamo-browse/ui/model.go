package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/commandctrl"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/dynamoitemview"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/dynamotableview"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/statusandprompt"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/tableselect"
)

type Model struct {
	tableReadController *controllers.TableReadController
	commandController   *commandctrl.CommandController
	statusAndPrompt *statusandprompt.StatusAndPrompt
	tableSelect *tableselect.Model

	root tea.Model
}

func NewModel(rc *controllers.TableReadController, cc *commandctrl.CommandController) Model {
	dtv := dynamotableview.New()
	div := dynamoitemview.New()
	statusAndPrompt := statusandprompt.New(layout.NewVBox(layout.LastChildFixedAt(17), dtv, div), "")
	tableSelect := tableselect.New(statusAndPrompt)

	root := layout.FullScreen(tableSelect)

	return Model{
		tableReadController: rc,
		commandController:   cc,
		statusAndPrompt: statusAndPrompt,
		tableSelect: tableSelect,
		root:                root,
	}
}

func (m Model) Init() tea.Cmd {
	return m.tableReadController.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.statusAndPrompt.InPrompt() && !m.tableSelect.Visible() {
			switch msg.String() {
			case "s":
				return m, m.tableReadController.Rescan()
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
