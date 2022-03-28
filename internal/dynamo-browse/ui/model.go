package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/dynamoitemview"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/dynamotableview"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/statusandprompt"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/tableselect"
)

type Model struct {
	tableReadController *controllers.TableReadController

	root tea.Model
}

func NewModel(rc *controllers.TableReadController) Model {
	dtv := dynamotableview.New(rc)
	div := dynamoitemview.New()

	m := statusandprompt.New(
		layout.NewVBox(layout.LastChildFixedAt(17), dtv, div),
		"Hello world",
	)
	root := layout.FullScreen(tableselect.New(m))

	return Model{
		tableReadController: rc,
		root:                root,
	}
}

func (m Model) Init() tea.Cmd {
	return m.tableReadController.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.root, cmd = m.root.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.root.View()
}
