package messageeditview

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/sqs-browse/controllers"
	"log"
)

type Model struct {
	w, h    int
	isShown bool

	child    tea.Model
	textArea textarea.Model
}

func New(child tea.Model) *Model {
	textArea := textarea.New()
	textArea.Placeholder = "ASA"
	textArea.SetValue("Hello?")
	textArea.Focus()

	return &Model{child: child, textArea: textArea}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.(type) {
	case controllers.EditMessage:
		m.isShown = true
		return m, cmd
	}

	if !m.isShown {
		m.child, cmd = m.child.Update(msg)
		return m, cmd
	}

	log.Printf("event %T", msg)
	m.textArea, cmd = m.textArea.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	if !m.isShown {
		return m.child.View()
	}

	return m.textArea.View()
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.w, m.h = w, h
	m.child = layout.Resize(m.child, w, h)
	m.textArea.SetWidth(w)
	m.textArea.SetHeight(h)
	return m
}
