package osstyle

import (
	"github.com/pkg/errors"
	"log"
	"os/exec"
	"strings"
)

const (
	errorMessageIndicatingInLightMode = `The domain/default pair of (kCFPreferencesAnyApplication, AppleInterfaceStyle) does not exist`
)

// Usage: https://stefan.sofa-rockers.org/2018/10/23/macos-dark-mode-terminal-vim/
func darwinGetOSColorScheme() ColorScheme {
	d, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			stdErr := string(exitErr.Stderr)

			if strings.Contains(stdErr, errorMessageIndicatingInLightMode) {
				log.Printf("error message indicates that macOS is in light mode")
				return ColorSchemeLightMode
			}

			log.Printf("cannot get current OS color scheme: %v - stderr: [%v]", err, stdErr)
		} else {
			log.Printf("cannot get current OS color scheme: %v", err)
		}

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
