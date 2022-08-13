package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/audax/internal/common/ui/commandctrl"
	"github.com/lmika/audax/internal/common/ui/logging"
	"github.com/lmika/audax/internal/common/ui/osstyle"
	"github.com/lmika/audax/internal/common/workspaces"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/lmika/audax/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/audax/internal/dynamo-browse/providers/workspacestore"
	"github.com/lmika/audax/internal/dynamo-browse/services/tables"
	workspaces_service "github.com/lmika/audax/internal/dynamo-browse/services/workspaces"
	"github.com/lmika/audax/internal/dynamo-browse/ui"
	"github.com/lmika/gopkgs/cli"
	"log"
	"net"
	"os"
)

func main() {
	var flagTable = flag.String("t", "", "dynamodb table name")
	var flagLocal = flag.String("local", "", "local endpoint")
	var flagDebug = flag.String("debug", "", "file to log debug messages")
	var flagWorkspace = flag.String("w", "", "workspace file")
	flag.Parse()

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		cli.Fatalf("cannot load AWS config: %v", err)
	}

	closeFn := logging.EnableLogging(*flagDebug)
	defer closeFn()

	wsManager := workspaces.New(workspaces.MetaInfo{Command: "dynamo-browse"})
	ws, err := wsManager.OpenOrCreate(*flagWorkspace)
	if err != nil {
		cli.Fatalf("cannot create workspace: %v", ws)
	}
	defer ws.Close()

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
	resultSetSnapshotStore := workspacestore.NewResultSetSnapshotStore(ws)

	tableService := tables.NewService(dynamoProvider)
	workspaceService := workspaces_service.NewService(resultSetSnapshotStore)

	state := controllers.NewState()
	tableReadController := controllers.NewTableReadController(state, tableService, workspaceService, *flagTable)
	tableWriteController := controllers.NewTableWriteController(state, tableService, tableReadController)

	commandController := commandctrl.NewCommandController()
	model := ui.NewModel(tableReadController, tableWriteController, commandController)

	// Pre-determine if layout has dark background.  This prevents calls for creating a list to hang.
	lipgloss.HasDarkBackground()

	p := tea.NewProgram(model, tea.WithAltScreen())

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
