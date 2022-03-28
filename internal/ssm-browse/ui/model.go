package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/statusandprompt"
	"github.com/lmika/awstools/internal/ssm-browse/controllers"
	"github.com/lmika/awstools/internal/ssm-browse/ui/ssmlist"
)

type Model struct {
	controller *controllers.SSMController

	root tea.Model
	ssmList *ssmlist.Model
}

func NewModel(controller *controllers.SSMController) Model {
	ssmList := ssmlist.New()
	root := layout.FullScreen(
		statusandprompt.New(ssmList, "Hello SSM"),
	)

	return Model{
		controller: controller,
		root: root,
		ssmList: ssmList,
	}
}


func (m Model) Init() tea.Cmd {
	return m.controller.Fetch()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case controllers.NewParameterListMsg:
		m.ssmList.SetParameters(msg.Parameters)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	newRoot, cmd := m.root.Update(msg)
	m.root = newRoot
	return m, cmd
}

func (m Model) View() string {
	return m.root.View()
}

