package controllers

import (
	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/slog-view/models"
	"github.com/lmika/audax/internal/slog-view/services/logreader"
	"golang.design/x/clipboard"
	"sync"
)

type LogFileController struct {
	logReader *logreader.Service

	// state
	mutex    *sync.Mutex
	filename string
	logFile  *models.LogFile
}

func NewLogFileController(logReader *logreader.Service, filename string) *LogFileController {
	return &LogFileController{
		logReader: logReader,
		filename:  filename,
		mutex:     new(sync.Mutex),
	}
}

func (lfc *LogFileController) ReadLogFile() tea.Cmd {
	return func() tea.Msg {
		logFile, err := lfc.logReader.Open(lfc.filename)
		if err != nil {
			return events.Error(err)
		}

		return NewLogFile(logFile)
	}
}

func (lfc *LogFileController) ViewLogLineFullScreen(line *models.LogLine) tea.Cmd {
	if line == nil {
		return nil
	}

	return func() tea.Msg {
		return ViewLogLineFullScreen(line)
	}
}

func (lfc *LogFileController) CopyLogLineToClipboard(line *models.LogLine) tea.Cmd {
	if line == nil {
		return nil
	}

	return func() tea.Msg {
		formattedJson, err := json.MarshalIndent(line.JSON, "", "   ")
		if err != nil {
			return events.Error(err)
		}

		clipboard.Write(clipboard.FmtText, formattedJson)

		return events.SetStatus("copied to clipboard")
	}
}
