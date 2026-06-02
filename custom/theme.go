// custom/theme.go
package custom

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// AppTheme overrides the default Fyne theme with the app's brand colours.
type AppTheme struct {
	fyne.Theme
}

// NewAppTheme returns the configured app theme.
func NewAppTheme() fyne.Theme {
	return &AppTheme{Theme: theme.DefaultTheme()}
}

// Color overrides selected colour tokens. All others fall back to DefaultTheme.
func (t *AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 37, G: 99, B: 235, A: 255} // blue-600
	case theme.ColorNameFocus:
		return color.NRGBA{R: 37, G: 99, B: 235, A: 180}
	}
	return t.Theme.Color(name, variant)
}
