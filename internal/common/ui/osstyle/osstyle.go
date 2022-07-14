package osstyle

type ColorScheme int

const (
	ColorSchemeUnknown ColorScheme = iota
	ColorSchemeLightMode
	ColorSchemeDarkMode
)

var getOSColorScheme func() ColorScheme = nil

func CurrentColorScheme() ColorScheme {
	if getOSColorScheme == nil {
		return ColorSchemeUnknown
	}
	return getOSColorScheme()
}
