// widgets/filter_bar.go
package widgets

import (
	"github.com/uchup07/fyne-jira-worklog-tracker/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FilterBar renders date range inputs plus Search and Cancel buttons.
type FilterBar struct {
	fs     *state.FilterState
	canvas fyne.CanvasObject
}

// NewFilterBar creates the filter bar. onSearch and onCancel are called when
// the respective buttons are clicked.
func NewFilterBar(fs *state.FilterState, onSearch, onCancel func()) *FilterBar {
	fb := &FilterBar{fs: fs}

	startEntry := widget.NewEntryWithData(fs.StartDate)
	startEntry.PlaceHolder = "YYYY-MM-DD"

	endEntry := widget.NewEntryWithData(fs.EndDate)
	endEntry.PlaceHolder = "YYYY-MM-DD"

	searchBtn := widget.NewButton("Search", onSearch)
	cancelBtn := widget.NewButton("Cancel", onCancel)

	fb.canvas = container.NewHBox(
		widget.NewLabel("From:"), startEntry,
		widget.NewLabel("To:"), endEntry,
		searchBtn,
		cancelBtn,
	)
	return fb
}

// Canvas returns the Fyne canvas object.
func (fb *FilterBar) Canvas() fyne.CanvasObject { return fb.canvas }
