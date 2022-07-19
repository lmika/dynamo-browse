package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/common/ui/commandctrl"
	"github.com/lmika/awstools/internal/common/ui/logging"
	"github.com/lmika/awstools/internal/common/ui/osstyle"
	"github.com/lmika/awstools/internal/common/workspaces"
	"log"
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

	_, _ = uiDispatcher, messageService

	commandController := commandctrl.NewCommandController()
	uiModel := ui.NewModel(commandController)

	p := tea.NewProgram(uiModel, tea.WithAltScreen())
	loopback.program = p

	bus.On("new-messages", func(m []*models.Message) { p.Send(ui.NewMessagesEvent(m)) })

	closeFn := logging.EnableLogging(*flagDebug)
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
