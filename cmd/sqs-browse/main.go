package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/lmika/awstools/internal/common/ui/logging"
	"github.com/lmika/awstools/internal/common/workspaces"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/dispatcher"
	"github.com/lmika/awstools/internal/sqs-browse/models"
	sqsprovider "github.com/lmika/awstools/internal/sqs-browse/providers/sqs"
	"github.com/lmika/awstools/internal/sqs-browse/providers/stores"
	"github.com/lmika/awstools/internal/sqs-browse/services/messages"
	"github.com/lmika/awstools/internal/sqs-browse/ui"
	"github.com/lmika/events"
	"github.com/lmika/gopkgs/cli"
)

func main() {
	//var flagQueue = flag.String("q", "", "queue to poll")
	//var flagTarget = flag.String("t", "", "target queue to push to")
	var flagDebug = flag.String("debug", "", "file to log debug messages")
	flag.Parse()

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		cli.Fatalf("cannot load AWS config: %v", err)
	}
	sqsClient := sqs.NewFromConfig(cfg)

	bus := events.New()

	wsManager := workspaces.New(workspaces.MetaInfo{
		Command: "sqs-browse",
	})
	ws, err := wsManager.CreateTemp()
	if err != nil {
		cli.Fatalf("cannot create workspace: %v", ws)
	}
	defer ws.Close()

	msgStore := stores.NewMessageStore(ws)
	sqsProvider := sqsprovider.NewProvider(sqsClient)

	messageService := messages.NewService(msgStore, sqsProvider)
	//pollService := pollmessage.NewService(msgStore, sqsProvider, *flagQueue, bus)

	//msgSendingHandlers := controllers.NewMessageSendingController(messageService, *flagTarget)

	loopback := &msgLoopback{}
	uiDispatcher := dispatcher.NewDispatcher(loopback)

	uiModel := ui.NewModel(uiDispatcher, msgSendingHandlers)
	p := tea.NewProgram(uiModel, tea.WithAltScreen())
	loopback.program = p

	bus.On("new-messages", func(m []*models.Message) { p.Send(ui.NewMessagesEvent(m)) })

	closeFn := logging.EnableLogging(*flagDebug)
	defer closeFn()

	//go func() {
	//	if err := pollService.Poll(context.Background()); err != nil {
	//		log.Printf("cannot start poller: %v", err)
	//	}
	//}()

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
