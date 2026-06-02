// screens/setup.go
package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Setup is the first-run configuration wizard.
// Full implementation comes in Plan 3.
type Setup struct {
	canvas fyne.CanvasObject
}

func NewSetup() *Setup {
	s := &Setup{}
	s.canvas = widget.NewLabel("Setup Wizard — full implementation in Plan 3")
	return s
}

func (s *Setup) Canvas() fyne.CanvasObject { return s.canvas }
