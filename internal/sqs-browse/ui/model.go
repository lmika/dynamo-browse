package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/common/ui/dispatcher"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/sqs-browse/controllers"
	"github.com/lmika/awstools/internal/sqs-browse/models"
	table "github.com/lmika/go-bubble-table"
)

var (
	activeHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#eac610"))

	inactiveHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#d1d1d1"))
)

type uiModel struct {
	table    table.Model
	viewport viewport.Model

	ready     bool
	tableRows []table.Row
	message   string

	pendingInput *events.PromptForInputMsg
	textInput    textinput.Model

	dispatcher         *dispatcher.Dispatcher
	msgSendingHandlers *controllers.MessageSendingController
}

func NewModel(dispatcher *dispatcher.Dispatcher, msgSendingHandlers *controllers.MessageSendingController) tea.Model {
	tbl := table.New(table.SimpleColumns{"seq", "message"}, 100, 20)
	rows := make([]table.Row, 0)
	tbl.SetRows(rows)

	textInput := textinput.New()

	model := uiModel{
		table:              tbl,
		tableRows:          rows,
		message:            "",
		textInput:          textInput,
		msgSendingHandlers: msgSendingHandlers,
		dispatcher:         dispatcher,
	}

	return model
}

func (m uiModel) Init() tea.Cmd {
	return nil
}

func (m *uiModel) updateViewportToSelectedMessage() {
	if message, ok := m.selectedMessage(); ok {
		// TODO: not all messages are JSON
		formattedJson := new(bytes.Buffer)
		if err := json.Indent(formattedJson, []byte(message.Data), "", "   "); err == nil {
			m.viewport.SetContent(formattedJson.String())
		} else {
			m.viewport.SetContent(message.Data)
		}
	} else {
		m.viewport.SetContent("(no message selected)")
	}
}

func (m uiModel) selectedMessage() (models.Message, bool) {
	if m.ready && len(m.tableRows) > 0 {
		if message, ok := m.table.SelectedRow().(messageTableRow); ok {
			return models.Message(message), true
		}
	}
	return models.Message{}, false
}

func (m uiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var textInputCommands tea.Cmd

	switch msg := msg.(type) {
	// Shared messages
	case events.ErrorMsg:
		m.message = "Error: " + msg.Error()
	case events.StatusMsg:
		m.message = string(msg)
	case events.PromptForInputMsg:
		// TODO
		//m.textInput.Focus()
		//m.textInput.SetValue("")
		//m.pendingInput = &msg

	// Local messages
	case NewMessagesEvent:
		for _, newMsg := range msg {
			m.tableRows = append(m.tableRows, messageTableRow(*newMsg))
		}
		m.table.SetRows(m.tableRows)
		m.updateViewportToSelectedMessage()

	case tea.WindowSizeMsg:
		fixedViewsHeight := lipgloss.Height(m.headerView()) + lipgloss.Height(m.splitterView()) + lipgloss.Height(m.footerView())

		if !m.ready {
			tableHeight := msg.Height / 2

			m.table.SetSize(msg.Width, tableHeight)
			m.viewport = viewport.New(msg.Width, msg.Height-tableHeight-fixedViewsHeight)
			m.viewport.SetContent("(no message selected)")
			m.ready = true
			log.Println("Viewport is now ready")
		} else {
			tableHeight := msg.Height / 2

			m.table.SetSize(msg.Width, tableHeight)
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - tableHeight - fixedViewsHeight
		}

		m.textInput.Width = msg.Width

		m.textInput, textInputCommands = m.textInput.Update(msg)
	case tea.KeyMsg:

		// If text input in focus, allow that to accept input messages
		if m.pendingInput != nil {
			switch msg.String() {
			case "ctrl+c", "esc":
				m.pendingInput = nil
			case "enter":
				//m.dispatcher.Start(uimodels.WithPromptValue(context.Background(), m.textInput.Value()), m.pendingInput.OnDone)
				m.pendingInput = nil
			default:
				m.textInput, textInputCommands = m.textInput.Update(msg)
			}
			break
		}

		// Normal focus
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "i":
			m.table.GoUp()
			m.updateViewportToSelectedMessage()
		case "down", "k":
			m.table.GoDown()
			m.updateViewportToSelectedMessage()

		// TODO: these should be moved somewhere else
		case "f":
			if selectedMessage, ok := m.selectedMessage(); ok {
				m.dispatcher.Start(context.Background(), m.msgSendingHandlers.ForwardMessage(selectedMessage))
			}
		}
	default:
		m.textInput, textInputCommands = m.textInput.Update(msg)
	}

	updatedTable, tableMsgs := m.table.Update(msg)
	updatedViewport, viewportMsgs := m.viewport.Update(msg)

	m.table = updatedTable
	m.viewport = updatedViewport

	return m, tea.Batch(textInputCommands, tableMsgs, viewportMsgs)
}

func (m uiModel) View() string {
	if !m.ready {
		return "Initializing"
	}

	if m.pendingInput != nil {
		return lipgloss.JoinVertical(lipgloss.Top,
			m.headerView(),
			m.table.View(),
			m.splitterView(),
			m.viewport.View(),
			m.textInput.View(),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		m.headerView(),
		m.table.View(),
		m.splitterView(),
		m.viewport.View(),
		m.footerView(),
	)
}

func (m uiModel) headerView() string {
	title := activeHeaderStyle.Render("Queue: XXX")
	line := activeHeaderStyle.Render(strings.Repeat(" ", max(0, m.viewport.Width-lipgloss.Width(title))))
	return lipgloss.JoinHorizontal(lipgloss.Left, title, line)
}

func (m uiModel) splitterView() string {
	title := inactiveHeaderStyle.Render("Message")
	line := inactiveHeaderStyle.Render(strings.Repeat(" ", max(0, m.viewport.Width-lipgloss.Width(title))))
	return lipgloss.JoinHorizontal(lipgloss.Left, title, line)
}

func (m uiModel) footerView() string {
	title := m.message
	line := strings.Repeat(" ", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Left, title, line)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
