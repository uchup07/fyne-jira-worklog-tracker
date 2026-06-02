// screens/report.go
package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Report shows the mapping/CR report.
// Full implementation comes in Plan 3.
type Report struct {
	canvas fyne.CanvasObject
}

func NewReport() *Report {
	r := &Report{}
	r.canvas = widget.NewLabel("Report — full implementation in Plan 3")
	return r
}

func (r *Report) Canvas() fyne.CanvasObject { return r.canvas }
