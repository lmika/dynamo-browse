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
	"github.com/lmika/awstools/internal/common/ui/osstyle"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/awstools/internal/dynamo-browse/ui"
	"github.com/lmika/gopkgs/cli"
	"log"
	"net"
	"os"
)

func main() {
	var flagTable = flag.String("t", "", "dynamodb table name")
	var flagLocal = flag.String("local", "", "local endpoint")
	var flagDebug = flag.String("debug", "", "file to log debug messages")
	flag.Parse()

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		cli.Fatalf("cannot load AWS config: %v", err)
	}

	var dynamoClient *dynamodb.Client
	if *flagLocal != "" {
		host, port, err := net.SplitHostPort(*flagLocal)
		if err != nil {
			cli.Fatalf("invalid address '%v': %v", *flagLocal, err)
		}
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "8000"
		}
		dynamoClient = dynamodb.NewFromConfig(cfg,
			dynamodb.WithEndpointResolver(dynamodb.EndpointResolverFromURL(fmt.Sprintf("http://%v:%v", host, port))))
	} else {
		dynamoClient = dynamodb.NewFromConfig(cfg)
	}

	dynamoProvider := dynamo.NewProvider(dynamoClient)

	tableService := tables.NewService(dynamoProvider)

	tableReadController := controllers.NewTableReadController(tableService, *flagTable)
	tableWriteController := controllers.NewTableWriteController(tableService, tableReadController)

	commandController := commandctrl.NewCommandController()
	model := ui.NewModel(tableReadController, tableWriteController, commandController)

	// Pre-determine if layout has dark background.  This prevents calls for creating a list to hang.
	lipgloss.HasDarkBackground()

	p := tea.NewProgram(model, tea.WithAltScreen())

	closeFn := logging.EnableLogging()
	defer closeFn()

	// Pre-determine if layout has dark background.  This prevents calls for creating a list to hang.
	if lipgloss.HasDarkBackground() {
		if colorScheme := osstyle.CurrentColorScheme(); colorScheme == osstyle.ColorSchemeLightMode {
			log.Printf("terminal reads dark but really in light mode")
			lipgloss.SetHasDarkBackground(true)
		} else {
			log.Printf("in dark background")
		}
	} else {
		if colorScheme := osstyle.CurrentColorScheme(); colorScheme == osstyle.ColorSchemeDarkMode {
			log.Printf("terminal reads light but really in dark mode")
			lipgloss.SetHasDarkBackground(true)
		} else {
			log.Printf("cannot detect system darkmode")
		}
	}

	log.Println("launching")
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
