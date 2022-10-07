package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/commandctrl"
	"github.com/lmika/audax/internal/common/ui/logging"
	"github.com/lmika/audax/internal/common/ui/osstyle"
	"github.com/lmika/audax/internal/common/workspaces"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/lmika/audax/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/audax/internal/dynamo-browse/providers/settingstore"
	"github.com/lmika/audax/internal/dynamo-browse/providers/workspacestore"
	"github.com/lmika/audax/internal/dynamo-browse/services/itemrenderer"
	"github.com/lmika/audax/internal/dynamo-browse/services/jobs"
	keybindings_service "github.com/lmika/audax/internal/dynamo-browse/services/keybindings"
	"github.com/lmika/audax/internal/dynamo-browse/services/tables"
	workspaces_service "github.com/lmika/audax/internal/dynamo-browse/services/workspaces"
	"github.com/lmika/audax/internal/dynamo-browse/ui"
	"github.com/lmika/audax/internal/dynamo-browse/ui/keybindings"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/styles"
	bus "github.com/lmika/events"
	"github.com/lmika/gopkgs/cli"
	"log"
	"net"
	"os"
)

func main() {
	var flagTable = flag.String("t", "", "dynamodb table name")
	var flagLocal = flag.String("local", "", "local endpoint")
	var flagDebug = flag.String("debug", "", "file to log debug messages")
	var flagRO = flag.Bool("ro", false, "enable readonly mode")
	var flagDefaultLimit = flag.Int("default-limit", 0, "default limit for queries and scans")
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

	eventBus := bus.New()

	uiStyles := styles.DefaultStyles
	dynamoProvider := dynamo.NewProvider(dynamoClient)
	resultSetSnapshotStore := workspacestore.NewResultSetSnapshotStore(ws)
	settingStore := settingstore.New(ws)

	if *flagRO {
		if err := settingStore.SetReadOnly(*flagRO); err != nil {
			cli.Fatalf("unable to set read-only mode: %v", err)
		}
	}
	if *flagDefaultLimit > 0 {
		if err := settingStore.SetDefaultLimit(*flagDefaultLimit); err != nil {
			cli.Fatalf("unable to set default limit: %v", err)
		}
	}

	tableService := tables.NewService(dynamoProvider, settingStore)
	workspaceService := workspaces_service.NewService(resultSetSnapshotStore)
	itemRendererService := itemrenderer.NewService(uiStyles.ItemView.FieldType, uiStyles.ItemView.MetaInfo)
	jobsService := jobs.NewService(eventBus)

	state := controllers.NewState()
	jobsController := controllers.NewJobsController(jobsService, eventBus)
	tableReadController := controllers.NewTableReadController(state, tableService, workspaceService, itemRendererService, jobsController, eventBus, *flagTable)
	tableWriteController := controllers.NewTableWriteController(state, tableService, jobsController, tableReadController, settingStore)
	columnsController := controllers.NewColumnsController(eventBus)
	exportController := controllers.NewExportController(state, columnsController)
	settingsController := controllers.NewSettingsController(settingStore)
	keyBindings := keybindings.Default()

	keyBindingService := keybindings_service.NewService(keyBindings)
	keyBindingController := controllers.NewKeyBindingController(keyBindingService)

	commandController := commandctrl.NewCommandController()

	model := ui.NewModel(
		tableReadController,
		tableWriteController,
		columnsController,
		exportController,
		settingsController,
		jobsController,
		itemRendererService,
		commandController,
		keyBindingController,
		keyBindings,
	)

	// Pre-determine if layout has dark background.  This prevents calls for creating a list to hang.
	osstyle.DetectCurrentScheme()

	p := tea.NewProgram(model, tea.WithAltScreen())

	jobsController.SetMessageSender(p.Send)

	log.Println("launching")
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
