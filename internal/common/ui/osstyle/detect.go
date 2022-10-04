package osstyle

import (
	"github.com/charmbracelet/lipgloss"
	"log"
)

func DetectCurrentScheme() {
	if lipgloss.HasDarkBackground() {
		if colorScheme := CurrentColorScheme(); colorScheme == ColorSchemeLightMode {
			log.Printf("terminal reads dark but really in light mode")
			lipgloss.SetHasDarkBackground(true)
		} else {
			log.Printf("in dark background")
		}
	} else {
		if colorScheme := CurrentColorScheme(); colorScheme == ColorSchemeDarkMode {
			log.Printf("terminal reads light but really in dark mode")
			lipgloss.SetHasDarkBackground(true)
		} else {
			log.Printf("cannot detect system darkmode")
		}
	}
}
