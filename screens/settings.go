// screens/settings.go
package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Settings shows Jira config, holidays, and language.
// Full implementation comes in Plan 4.
type Settings struct {
	canvas fyne.CanvasObject
}

func NewSettings() *Settings {
	s := &Settings{}
	s.canvas = widget.NewLabel("Settings — full implementation in Plan 4")
	return s
}

func (s *Settings) Canvas() fyne.CanvasObject { return s.canvas }
