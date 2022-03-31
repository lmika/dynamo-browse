package loglines

import (
	table "github.com/calyptia/go-bubble-table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/frame"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/slog-view/models"
	"path/filepath"
)

var (
	activeHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#9c9c9c"))
)

type Model struct {
	frameTitle frame.FrameTitle
	table      table.Model

	logFile *models.LogFile

	w, h int
}

func New() *Model {
	frameTitle := frame.NewFrameTitle("File: ", true, activeHeaderStyle)
	table := table.New([]string{"level", "error", "message"}, 0, 0)

	return &Model{
		frameTitle: frameTitle,
		table:      table,
	}
}

func (m *Model) SetLogFile(newLogFile *models.LogFile) {
	m.logFile = newLogFile
	m.frameTitle.SetTitle("File: " + filepath.Base(newLogFile.Filename))

	cols := []string{"level", "error", "message"}

	newTbl := table.New(cols, m.w, m.h-m.frameTitle.HeaderHeight())
	newRows := make([]table.Row, len(newLogFile.Lines))
	for i, r := range newLogFile.Lines {
		newRows[i] = itemTableRow{r}
	}
	newTbl.SetRows(newRows)

	m.table = newTbl
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "i", "up":
			m.table.GoUp()
			return m, m.emitNewSelectedParameter()
		case "k", "down":
			m.table.GoDown()
			return m, m.emitNewSelectedParameter()
		}
		//m.table, cmd = m.table.Update(msg)
		//return m, cmd
	}
	return m, nil
}

func (m *Model) SelectedLogLine() *models.LogLine {
	if row, ok := m.table.SelectedRow().(itemTableRow); ok {
		return &(row.item)
	}
	return nil
}

func (m *Model) emitNewSelectedParameter() tea.Cmd {
	return func() tea.Msg {
		selectedLogLine := m.SelectedLogLine()
		if selectedLogLine != nil {
			return NewLogLineSelected(selectedLogLine)
		}

		return nil
	}
}

func (m *Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, m.frameTitle.View(), m.table.View())
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.w, m.h = w, h
	m.frameTitle.Resize(w, h)
	m.table.SetSize(w, h-m.frameTitle.HeaderHeight())
	return m
}
