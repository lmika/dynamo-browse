package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/commandctrl"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/statusandprompt"
	"github.com/lmika/audax/internal/slog-view/controllers"
	"github.com/lmika/audax/internal/slog-view/styles"
	"github.com/lmika/audax/internal/slog-view/ui/fullviewlinedetails"
	"github.com/lmika/audax/internal/slog-view/ui/linedetails"
	"github.com/lmika/audax/internal/slog-view/ui/loglines"
)

type Model struct {
	controller    *controllers.LogFileController
	cmdController *commandctrl.CommandController

	root                tea.Model
	logLines            *loglines.Model
	lineDetails         *linedetails.Model
	statusAndPrompt     *statusandprompt.StatusAndPrompt
	fullViewLineDetails *fullviewlinedetails.Model
}

func NewModel(controller *controllers.LogFileController, cmdController *commandctrl.CommandController) Model {
	defaultStyles := styles.DefaultStyles
	logLines := loglines.New(defaultStyles.Frames)
	lineDetails := linedetails.New(defaultStyles.Frames)
	box := layout.NewVBox(layout.LastChildFixedAt(17), logLines, lineDetails)
	fullViewLineDetails := fullviewlinedetails.NewModel(box, defaultStyles.Frames)
	statusAndPrompt := statusandprompt.New(fullViewLineDetails, "", defaultStyles.StatusAndPrompt)

	root := layout.FullScreen(statusAndPrompt)

	return Model{
		controller:          controller,
		cmdController:       cmdController,
		root:                root,
		statusAndPrompt:     statusAndPrompt,
		logLines:            logLines,
		lineDetails:         lineDetails,
		fullViewLineDetails: fullViewLineDetails,
	}
}

func (m Model) Init() tea.Cmd {
	return m.controller.ReadLogFile()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case controllers.NewLogFile:
		m.logLines.SetLogFile(msg)
	case controllers.ViewLogLineFullScreen:
		m.fullViewLineDetails.ViewItem(msg)
	case loglines.NewLogLineSelected:
		m.lineDetails.SetSelectedItem(msg)

	case tea.KeyMsg:
		if !m.statusAndPrompt.InPrompt() {
			switch msg.String() {
			// TEMP
			case ":":
				return m, func() tea.Msg { return m.cmdController.Prompt() }
			case "w":
				return m, m.controller.ViewLogLineFullScreen(m.logLines.SelectedLogLine())
			case "c":
				return m, m.controller.CopyLogLineToClipboard(m.logLines.SelectedLogLine())
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
