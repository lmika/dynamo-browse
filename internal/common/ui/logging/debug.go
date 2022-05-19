package logging

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func EnableLogging() (closeFn func()) {
	tempFile, err := os.CreateTemp("", "debug.log")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	tempFile.Close()

	f, err := tea.LogToFile(tempFile.Name(), "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	return func() {
		f.Close()
	}
}
