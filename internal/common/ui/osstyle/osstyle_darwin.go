package osstyle

import (
	"log"
	"os/exec"
)

// Usage: https://stefan.sofa-rockers.org/2018/10/23/macos-dark-mode-terminal-vim/
func darwinGetOSColorScheme() ColorScheme {
	d, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
	if err != nil {
		log.Printf("cannot get current OS color scheme: %v", err)
		return ColorSchemeUnknown
	}

	switch string(d) {
	case "Dark\n":
		return ColorSchemeDarkMode
	case "Light\n":
		return ColorSchemeLightMode
	}
	return ColorSchemeUnknown
}

func init() {
	getOSColorScheme = darwinGetOSColorScheme
}
