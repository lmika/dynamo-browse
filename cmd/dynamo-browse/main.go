package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/common/ui/commandctrl"
	"github.com/lmika/awstools/internal/common/ui/dispatcher"
	"github.com/lmika/awstools/internal/common/ui/uimodels"
	"github.com/lmika/awstools/internal/dynamo-browse/controllers"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/awstools/internal/dynamo-browse/ui"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/modal"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/tableselect"
	"github.com/lmika/gopkgs/cli"
)

func main() {
	var flagTable = flag.String("t", "", "dynamodb table name")
	var flagLocal = flag.Bool("local", false, "local endpoint")
	flag.Parse()

	ctx := context.Background()

	// TEMP
	cfg, err := config.LoadDefaultConfig(ctx)

	// END TEMP
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
		// "scan": tableReadController.Scan(),
		"rw":  tableWriteController.ToggleReadWrite(),
		"dup": tableWriteController.Duplicate(),
	})

	_ = uiDispatcher
	_ = commandController

	// uiModel := ui.NewModel(uiDispatcher, commandController, tableReadController, tableWriteController)

	// TEMP
	// _ = uiModel
	// END TEMP

	/*
		var model tea.Model = statusandprompt.New(
			layout.NewVBox(
				layout.LastChildFixedAt(11),
				dynamotableview.New(tableReadController),
				dynamoitemview.New(),
			),
			"Hello world",
		)
		model = layout.FullScreen(tableselect.New(model))
	*/
	model := ui.NewModel(tableReadController)

	// Pre-determine if layout has dark background.  This prevents calls for creating a list to hang.
	lipgloss.HasDarkBackground()

	//frameSet := frameset.New([]frameset.Frame{
	//	{
	//		Header: "Frame 1",
	//		Model:  newTestModel("this is model 1"),
	//	},
	//	{
	//		Header: "Frame 2",
	//		Model:  newTestModel("this is model 2"),
	//	},
	//})
	//
	//modal := modal.New(frameSet)

	p := tea.NewProgram(model, tea.WithAltScreen())
	//loopback.program = p

	// TEMP -- profiling
	//cf, err := os.Create("trace.out")
	//if err != nil {
	//	log.Fatal("could not create CPU profile: ", err)
	//}
	//defer cf.Close() // error handling omitted for example
	//if err := trace.Start(cf); err != nil {
	//	log.Fatal("could not start CPU profile: ", err)
	//}
	//defer trace.Stop()
	// END TEMP

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	log.Println("launching")
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

func newTestModel(descr string) tea.Model {
	return teamodels.TestModel{
		Message: descr,
		OnKeyPressed: func(k string) tea.Cmd {
			log.Println("got key press: " + k)
			if k == "enter" {
				return tea.Batch(
					tableselect.IndicateLoadingTables(),
					tea.Sequentially(
						func() tea.Msg {
							<-time.After(2 * time.Second)
							return nil
						},
						tableselect.ShowTableSelect(func(n string) tea.Cmd {
							// return statusandprompt.SetStatus("New table = " + n)
							return nil
						}),
					),
				)
			} else if k == "k" {
				return modal.PopMode
			}
			return nil
		},
	}
}
