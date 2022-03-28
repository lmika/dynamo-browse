package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/common/ui/logging"
	"github.com/lmika/awstools/internal/ssm-browse/controllers"
	"github.com/lmika/awstools/internal/ssm-browse/providers/awsssm"
	"github.com/lmika/awstools/internal/ssm-browse/services/ssmparameters"
	"github.com/lmika/awstools/internal/ssm-browse/ui"
	"github.com/lmika/gopkgs/cli"
	"os"
)

func main() {
	// Pre-determine if layout has dark background.  This prevents calls for creating a list to hang.
	lipgloss.HasDarkBackground()

	closeFn := logging.EnableLogging()
	defer closeFn()

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		cli.Fatalf("cannot load AWS config: %v", err)
	}
	ssmClient := ssm.NewFromConfig(cfg)

	provider := awsssm.NewProvider(ssmClient)
	service := ssmparameters.NewService(provider)

	ctrl := controllers.New(service)
	model := ui.NewModel(ctrl)

	p := tea.NewProgram(model, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
