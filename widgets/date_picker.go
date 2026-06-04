// widgets/date_picker.go
package widgets

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// DatePicker is a button that opens a calendar popup to select a date.
// It reads and writes a binding.String in "2006-01-02" format.
type DatePicker struct {
	b      binding.String
	btn    *widget.Button
	window fyne.Window
}

// NewDatePicker creates a DatePicker bound to b.
func NewDatePicker(b binding.String, window fyne.Window) *DatePicker {
	dp := &DatePicker{b: b, window: window}
	val, _ := b.Get()
	dp.btn = widget.NewButton(val, dp.open)
	b.AddListener(binding.NewDataListener(func() {
		v, _ := b.Get()
		fyne.Do(func() { dp.btn.SetText(v) })
	}))
	return dp
}

// Canvas returns the trigger button.
func (dp *DatePicker) Canvas() fyne.CanvasObject { return dp.btn }

func (dp *DatePicker) open() {
	val, _ := dp.b.Get()
	initial := time.Now()
	if t, err := time.Parse("2006-01-02", val); err == nil {
		initial = t
	}

	var cp *calPop
	cp = newCalPop(initial, func(t time.Time) {
		dp.b.Set(t.Format("2006-01-02"))
		cp.dlg.Hide()
	}, dp.window)
	cp.dlg.Show()
}

// calPop is an unexported calendar popup dialog.
type calPop struct {
	dlg        dialog.Dialog
	monthLabel *widget.Label
	grid       *fyne.Container
	viewing    time.Time
	selected   time.Time
	onSelect   func(time.Time)
}

func newCalPop(initial time.Time, onSelect func(time.Time), w fyne.Window) *calPop {
	cp := &calPop{
		viewing:  initial,
		selected: initial,
		onSelect: onSelect,
	}

	cp.monthLabel = widget.NewLabel("")
	cp.monthLabel.Alignment = fyne.TextAlignCenter

	prevBtn := widget.NewButton("<", func() {
		cp.viewing = cp.viewing.AddDate(0, -1, 0)
		cp.refresh()
	})
	nextBtn := widget.NewButton(">", func() {
		cp.viewing = cp.viewing.AddDate(0, 1, 0)
		cp.refresh()
	})

	header := container.NewBorder(nil, nil, prevBtn, nextBtn, cp.monthLabel)

	dayHdr := container.NewGridWithColumns(7)
	for _, d := range []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"} {
		dayHdr.Add(widget.NewLabelWithStyle(d, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	}

	cp.grid = container.NewGridWithColumns(7)
	cp.refresh()

	content := container.NewVBox(header, dayHdr, cp.grid)
	cp.dlg = dialog.NewCustom("Select Date", "Cancel", content, w)
	return cp
}

func (cp *calPop) refresh() {
	cp.monthLabel.SetText(cp.viewing.Format("January 2006"))

	cp.grid.Objects = nil

	first := time.Date(cp.viewing.Year(), cp.viewing.Month(), 1, 0, 0, 0, 0, time.Local)
	// Monday-first offset: Mon=0 … Sun=6
	offset := (int(first.Weekday()) - 1 + 7) % 7

	daysInMonth := time.Date(cp.viewing.Year(), cp.viewing.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()

	today := time.Now().Truncate(24 * time.Hour)
	selDay := time.Date(cp.selected.Year(), cp.selected.Month(), cp.selected.Day(), 0, 0, 0, 0, time.Local)

	for i := 0; i < offset; i++ {
		cp.grid.Add(widget.NewLabel(""))
	}

	for day := 1; day <= daysInMonth; day++ {
		day := day
		t := time.Date(cp.viewing.Year(), cp.viewing.Month(), day, 0, 0, 0, 0, time.Local)
		btn := widget.NewButton(fmt.Sprintf("%d", day), func() { cp.onSelect(t) })
		switch {
		case t.Equal(selDay):
			btn.Importance = widget.HighImportance
		case t.Equal(today):
			btn.Importance = widget.MediumImportance
		default:
			btn.Importance = widget.LowImportance
		}
		cp.grid.Add(btn)
	}

	cp.grid.Refresh()
}
