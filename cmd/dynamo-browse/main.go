package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/commandctrl"
	"github.com/lmika/awstools/internal/common/ui/dispatcher"
	"github.com/lmika/awstools/internal/common/ui/uimodels"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/awstools/internal/dynamo-browse/ui"
	"github.com/lmika/gopkgs/cli"
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

	loopback := &msgLoopback{}
	uiDispatcher := dispatcher.NewDispatcher(loopback)

	tableReadController := controllers.NewTableReadController(tableService, *flagTable)
	tableWriteController := controllers.NewTableWriteController(tableService, tableReadController, *flagTable)

	commandController := commandctrl.NewCommandController(map[string]uimodels.Operation{
		"scan": tableReadController.Scan(),
		"rw":   tableWriteController.ToggleReadWrite(),
		"dup":  tableWriteController.Duplicate(),
	})

	uiModel := ui.NewModel(uiDispatcher, commandController, tableReadController, tableWriteController)
	p := tea.NewProgram(uiModel, tea.WithAltScreen())
	loopback.program = p

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

type msgLoopback struct {
	program *tea.Program
}

func (m *msgLoopback) Send(msg tea.Msg) {
	m.program.Send(msg)
}
