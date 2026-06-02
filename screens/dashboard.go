// screens/dashboard.go
package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Dashboard is the main worklog-viewing screen.
// Full implementation comes in Plan 3.
type Dashboard struct {
	canvas fyne.CanvasObject
}

func NewDashboard() *Dashboard {
	d := &Dashboard{}
	d.canvas = widget.NewLabel("Dashboard — full implementation in Plan 3")
	return d
}

func (d *Dashboard) Canvas() fyne.CanvasObject { return d.canvas }
