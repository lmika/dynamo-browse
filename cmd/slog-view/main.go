package main

import (
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/common/ui/commandctrl"
	"github.com/lmika/awstools/internal/common/ui/logging"
	"github.com/lmika/awstools/internal/slog-view/services/logreader"
	"github.com/lmika/awstools/internal/slog-view/controllers"
	"github.com/lmika/awstools/internal/slog-view/ui"
	"github.com/lmika/gopkgs/cli"
	"os"
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		cli.Fatal("usage: slog-view LOGFILE")
	}

	// Pre-determine if layout has dark background.  This prevents calls for creating a list to hang.
	lipgloss.HasDarkBackground()

	closeFn := logging.EnableLogging()
	defer closeFn()

	service := logreader.NewService()

	ctrl := controllers.NewLogFileController(service, flag.Arg(0))

	cmdController := commandctrl.NewCommandController()
	//cmdController.AddCommands(&commandctrl.CommandContext{
	//	Commands: map[string]commandctrl.Command{
	//		"cd": func(args []string) tea.Cmd {
	//			return ctrl.ChangePrefix(args[0])
	//		},
	//	},
	//})

	model := ui.NewModel(ctrl, cmdController)

	p := tea.NewProgram(model, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}