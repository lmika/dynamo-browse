package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/commandctrl"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/statusandprompt"
	"github.com/lmika/audax/internal/ssm-browse/controllers"
	"github.com/lmika/audax/internal/ssm-browse/styles"
	"github.com/lmika/audax/internal/ssm-browse/ui/ssmdetails"
	"github.com/lmika/audax/internal/ssm-browse/ui/ssmlist"
	"github.com/pkg/errors"
)

type Model struct {
	cmdController   *commandctrl.CommandController
	controller      *controllers.SSMController
	statusAndPrompt *statusandprompt.StatusAndPrompt

	root       tea.Model
	ssmList    *ssmlist.Model
	ssmDetails *ssmdetails.Model
}

func NewModel(controller *controllers.SSMController, cmdController *commandctrl.CommandController) Model {
	defaultStyles := styles.DefaultStyles
	ssmList := ssmlist.New(defaultStyles.Frames)
	ssmdDetails := ssmdetails.New(defaultStyles.Frames)
	statusAndPrompt := statusandprompt.New(
		layout.NewVBox(layout.LastChildFixedAt(17), ssmList, ssmdDetails), "", defaultStyles.StatusAndPrompt)

	cmdController.AddCommands(&commandctrl.CommandContext{
		Commands: map[string]commandctrl.Command{
			"clone": func(args []string) tea.Msg {
				if currentParam := ssmList.CurrentParameter(); currentParam != nil {
					return controller.Clone(*currentParam)
				}
				return events.Error(errors.New("no parameter selected"))
			},
			"delete": func(args []string) tea.Msg {
				if currentParam := ssmList.CurrentParameter(); currentParam != nil {
					return controller.DeleteParameter(*currentParam)
				}
				return events.Error(errors.New("no parameter selected"))
			},
		},
	})

	root := layout.FullScreen(statusAndPrompt)

	return Model{
		controller:      controller,
		cmdController:   cmdController,
		root:            root,
		statusAndPrompt: statusAndPrompt,
		ssmList:         ssmList,
		ssmDetails:      ssmdDetails,
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
	case ssmlist.NewSSMParameterSelected:
		m.ssmDetails.SetSelectedItem(msg)
	case tea.KeyMsg:
		if !m.statusAndPrompt.InPrompt() {
			switch msg.String() {
			// TEMP
			case ":":
				return m, func() tea.Msg { return m.cmdController.Prompt() }
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
