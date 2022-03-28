package ssmlist

import (
	table "github.com/calyptia/go-bubble-table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/ssm-browse/models"
)

type Model struct {
	frameTitle frame.FrameTitle
	table      table.Model

	parameters *models.SSMParameters

	w, h int
}

func New() *Model {
	frameTitle := frame.NewFrameTitle("SSM: /", true)
	table := table.New([]string{"name", "type", "value"}, 0, 0)

	return &Model{
		frameTitle: frameTitle,
		table: table,
	}
}

func (m *Model) SetPrefix(newPrefix string) {
	m.frameTitle.SetTitle("SSM: " + newPrefix)
}

func (m *Model) SetParameters(parameters *models.SSMParameters) {
	m.parameters = parameters
	cols := []string{"name", "type", "value"}

	newTbl := table.New(cols, m.w, m.h-m.frameTitle.HeaderHeight())
	newRows := make([]table.Row, len(parameters.Items))
	for i, r := range parameters.Items {
		newRows[i] = itemTableRow{r}
	}
	newTbl.SetRows(newRows)

	m.table = newTbl
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, m.frameTitle.View(), m.table.View())
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.w, m.h = w, h
	m.frameTitle.Resize(w, h)
	m.table.SetSize(w, h - m.frameTitle.HeaderHeight())
	return m
}

