// widgets/progress_bar.go
package widgets

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

// SearchProgress displays a status label above a progress bar.
// Update it from any goroutine via the Value and Text bindings — thread-safe.
type SearchProgress struct {
	Value  binding.Float  // 0.0–1.0
	Text   binding.String // status message
	canvas fyne.CanvasObject
}

// NewSearchProgress creates the widget.
func NewSearchProgress() *SearchProgress {
	p := &SearchProgress{
		Value: binding.NewFloat(),
		Text:  binding.NewString(),
	}

	bar := widget.NewProgressBarWithData(p.Value)
	label := widget.NewLabelWithData(p.Text)
	label.Alignment = fyne.TextAlignCenter

	p.canvas = container.NewVBox(label, bar)
	return p
}

// Canvas returns the Fyne canvas object to embed in a screen.
func (p *SearchProgress) Canvas() fyne.CanvasObject { return p.canvas }

// Reset clears the progress bar and label.
func (p *SearchProgress) Reset() {
	p.Value.Set(0)
	p.Text.Set("")
}

// SetSearching updates the UI for the "searching" phase.
func (p *SearchProgress) SetSearching(pages, found int) {
	p.Value.Set(float64(pages) * 0.03)
	p.Text.Set(fmt.Sprintf("Searching... %d issues found", found))
}

// SetProcessing updates the UI for the "processing" phase.
func (p *SearchProgress) SetProcessing(done, total int) {
	ratio := 0.0
	if total > 0 {
		ratio = float64(done) / float64(total)
	}
	p.Value.Set(0.10 + ratio*0.85)
	p.Text.Set(fmt.Sprintf("Processing %d / %d", done, total))
}

// SetFinalizing updates the UI for the final aggregation step.
func (p *SearchProgress) SetFinalizing() {
	p.Value.Set(0.95)
	p.Text.Set("Finalizing results...")
}

// SetDone clears the bar after a search completes.
func (p *SearchProgress) SetDone() {
	p.Value.Set(1.0)
	p.Text.Set("")
}
