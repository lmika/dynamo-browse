package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/commandctrl"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/statusandprompt"
	"github.com/lmika/awstools/internal/ssm-browse/controllers"
	"github.com/lmika/awstools/internal/ssm-browse/ui/ssmlist"
)

type Model struct {
	cmdController   *commandctrl.CommandController
	controller      *controllers.SSMController
	statusAndPrompt *statusandprompt.StatusAndPrompt

	root    tea.Model
	ssmList *ssmlist.Model
}

func NewModel(controller *controllers.SSMController, cmdController *commandctrl.CommandController) Model {
	ssmList := ssmlist.New()
	statusAndPrompt := statusandprompt.New(ssmList, "Hello SSM")

	root := layout.FullScreen(statusAndPrompt)

	return Model{
		controller:      controller,
		cmdController:   cmdController,
		root:            root,
		statusAndPrompt: statusAndPrompt,
		ssmList:         ssmList,
	}
}

func (m Model) Init() tea.Cmd {
	return m.controller.Fetch()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case controllers.NewParameterListMsg:
		m.ssmList.SetPrefix(msg.Prefix)
		m.ssmList.SetParameters(msg.Parameters)
	case tea.KeyMsg:
		if !m.statusAndPrompt.InPrompt() {
			switch msg.String() {
			// TEMP
			case ":":
				return m, m.cmdController.Prompt()
			// END TEMP

			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}
	}

	newRoot, cmd := m.root.Update(msg)
	m.root = newRoot
	return m, cmd
}

func (m Model) View() string {
	return m.root.View()
}
