package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/common/ui/commandctrl"
	"github.com/lmika/awstools/internal/common/ui/logging"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/awstools/internal/dynamo-browse/ui"
	"github.com/lmika/gopkgs/cli"
	"log"
	"os"
)

func main() {
	var flagTable = flag.String("t", "", "dynamodb table name")
	var flagLocal = flag.Bool("local", false, "local endpoint")
	flag.Parse()

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		cli.Fatalf("cannot load AWS config: %v", err)
	}

	var dynamoClient *dynamodb.Client
	if *flagLocal {
		dynamoClient = dynamodb.NewFromConfig(cfg,
			dynamodb.WithEndpointResolver(dynamodb.EndpointResolverFromURL("http://localhost:8000")))
	} else {
		dynamoClient = dynamodb.NewFromConfig(cfg)
	}

	dynamoProvider := dynamo.NewProvider(dynamoClient)

	tableService := tables.NewService(dynamoProvider)

	tableReadController := controllers.NewTableReadController(tableService, *flagTable)
	tableWriteController := controllers.NewTableWriteController(tableService, tableReadController, *flagTable)
	_ = tableWriteController

	commandController := commandctrl.NewCommandController()
	commandController.AddCommands(&commandctrl.CommandContext{
		Commands: map[string]commandctrl.Command{
			"q": commandctrl.NoArgCommand(tea.Quit),
			"table": func(args []string) tea.Cmd {
				if len(args) == 0 {
					return tableReadController.ListTables()
				} else {
					return tableReadController.ScanTable(args[0])
				}
			},
		},
	})

	model := ui.NewModel(tableReadController, commandController)

	// Pre-determine if layout has dark background.  This prevents calls for creating a list to hang.
	lipgloss.HasDarkBackground()

	p := tea.NewProgram(model, tea.WithAltScreen())

	closeFn := logging.EnableLogging()
	defer closeFn()

	log.Println("launching")
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
