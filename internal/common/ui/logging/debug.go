package logging

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func EnableLogging(logFile string) (closeFn func()) {
	if logFile == "" {
		tempFile, err := os.CreateTemp("", "debug.log")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		tempFile.Close()
		logFile = tempFile.Name()
	}

	f, err := tea.LogToFile(logFile, "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	return func() {
		f.Close()
	}
}
