package logging

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func EnableLogging() (closeFn func()) {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	return func() {
		f.Close()
	}
}
