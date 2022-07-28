package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/dispatcher"
	"github.com/lmika/audax/internal/sqs-browse/controllers"
	"github.com/lmika/audax/internal/sqs-browse/models"
	sqsprovider "github.com/lmika/audax/internal/sqs-browse/providers/sqs"
	"github.com/lmika/audax/internal/sqs-browse/providers/stormstore"
	"github.com/lmika/audax/internal/sqs-browse/services/messages"
	"github.com/lmika/audax/internal/sqs-browse/services/pollmessage"
	"github.com/lmika/audax/internal/sqs-browse/ui"
	"github.com/lmika/events"
	"github.com/lmika/gopkgs/cli"
)

func main() {
	var flagQueue = flag.String("q", "", "queue to poll")
	var flagTarget = flag.String("t", "", "target queue to push to")
	flag.Parse()

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		cli.Fatalf("cannot load AWS config: %v", err)
	}
	sqsClient := sqs.NewFromConfig(cfg)

	bus := events.New()

	workspaceFile, err := os.CreateTemp("", "sqs-browse*.workspace")
	if err != nil {
		cli.Fatalf("cannot create workspace file: %v", err)
	}
	workspaceFile.Close() // We just need the filename

	msgStore, err := stormstore.NewStore(workspaceFile.Name())
	if err != nil {
		cli.Fatalf("cannot open workspace: %v", err)
	}
	defer msgStore.Close()

	sqsProvider := sqsprovider.NewProvider(sqsClient)

	messageService := messages.NewService(sqsProvider)
	pollService := pollmessage.NewService(msgStore, sqsProvider, *flagQueue, bus)

	msgSendingHandlers := controllers.NewMessageSendingController(messageService, *flagTarget)

	loopback := &msgLoopback{}
	uiDispatcher := dispatcher.NewDispatcher(loopback)

	uiModel := ui.NewModel(uiDispatcher, msgSendingHandlers)
	p := tea.NewProgram(uiModel, tea.WithAltScreen())
	loopback.program = p

	bus.On("new-messages", func(m []*models.Message) { p.Send(ui.NewMessagesEvent(m)) })

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	log.Printf("workspace file: %v", workspaceFile.Name())

	go func() {
		if err := pollService.Poll(context.Background()); err != nil {
			log.Printf("cannot start poller: %v", err)
		}
	}()

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
