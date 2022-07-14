package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/common/ui/commandctrl"
	"github.com/lmika/awstools/internal/common/ui/logging"
	"github.com/lmika/awstools/internal/ssm-browse/controllers"
	"github.com/lmika/awstools/internal/ssm-browse/providers/awsssm"
	"github.com/lmika/awstools/internal/ssm-browse/services/ssmparameters"
	"github.com/lmika/awstools/internal/ssm-browse/ui"
	"github.com/lmika/gopkgs/cli"
	"os"
)

func main() {
	var flagLocal = flag.Bool("local", false, "local endpoint")
	var flagDebug = flag.String("debug", "", "file to log debug messages")
	flag.Parse()

	// Pre-determine if layout has dark background.  This prevents calls for creating a list to hang.
	lipgloss.HasDarkBackground()

	closeFn := logging.EnableLogging(*flagDebug)
	defer closeFn()

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		cli.Fatalf("cannot load AWS config: %v", err)
	}

	var ssmClient *ssm.Client
	if *flagLocal {
		ssmClient = ssm.NewFromConfig(cfg,
			ssm.WithEndpointResolver(ssm.EndpointResolverFromURL("http://localhost:4566")))
	} else {
		ssmClient = ssm.NewFromConfig(cfg)
	}

	provider := awsssm.NewProvider(ssmClient)
	service := ssmparameters.NewService(provider)

	ctrl := controllers.New(service)

	cmdController := commandctrl.NewCommandController()
	cmdController.AddCommands(&commandctrl.CommandContext{
		Commands: map[string]commandctrl.Command{
			"cd": func(args []string) tea.Cmd {
				return ctrl.ChangePrefix(args[0])
			},
		},
	})

	model := ui.NewModel(ctrl, cmdController)

	p := tea.NewProgram(model, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
