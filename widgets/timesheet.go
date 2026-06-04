// widgets/timesheet.go
package widgets

import (
	"fmt"
	"image/color"
	"time"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/uchup07/fyne-jira-worklog-tracker/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

// Timesheet renders a monthly heatmap of hours logged per day with month navigation.
type Timesheet struct {
	ws           *state.WorklogState
	dayTotals    map[string]int // aggregated from the current search result
	viewingMonth time.Time      // month currently shown in the grid
	monthLabel   *widget.Label
	calWrap      *fyne.Container // swapped on every month change
	canvas       fyne.CanvasObject
}

// NewTimesheet creates a timesheet bound to ws.Groups.
func NewTimesheet(ws *state.WorklogState) *Timesheet {
	ts := &Timesheet{
		ws:           ws,
		dayTotals:    map[string]int{},
		viewingMonth: time.Now(),
	}

	ts.monthLabel = widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	ts.calWrap = container.NewStack()

	prevBtn := widget.NewButton("< Prev", func() {
		ts.viewingMonth = ts.viewingMonth.AddDate(0, -1, 0)
		ts.rebuildCalendar()
	})
	nextBtn := widget.NewButton("Next >", func() {
		ts.viewingMonth = ts.viewingMonth.AddDate(0, 1, 0)
		ts.rebuildCalendar()
	})

	nav := container.NewBorder(nil, nil, prevBtn, nextBtn, ts.monthLabel)
	ts.canvas = container.NewBorder(nav, nil, nil, nil, ts.calWrap)

	ts.rebuildCalendar()

	// Groups.Set is called inside fyne.Do, so this listener fires on the main goroutine.
	ws.Groups.AddListener(binding.NewDataListener(func() {
		ts.loadData()
	}))

	return ts
}

// Canvas returns the Fyne canvas object.
func (ts *Timesheet) Canvas() fyne.CanvasObject { return ts.canvas }

// loadData recomputes dayTotals from the current groups and resets to the search-start month.
func (ts *Timesheet) loadData() {
	dayTotals := map[string]int{}
	for i := 0; i < ts.ws.Groups.Length(); i++ {
		val, err := ts.ws.Groups.GetValue(i)
		if err != nil {
			continue
		}
		group, ok := val.(jira.WorklogGroup)
		if !ok {
			continue
		}
		for _, item := range group.Items {
			dayTotals[item.Started.Format("2006-01-02")] += item.TimeSpentSeconds
		}
	}
	ts.dayTotals = dayTotals

	monthStr, _ := ts.ws.SearchStart.Get()
	if t, err := time.Parse("2006-01-02", monthStr); err == nil {
		ts.viewingMonth = t
	} else {
		ts.viewingMonth = time.Now()
	}

	ts.rebuildCalendar()
}

// rebuildCalendar replaces the grid for viewingMonth using the cached dayTotals.
func (ts *Timesheet) rebuildCalendar() {
	ts.monthLabel.SetText(ts.viewingMonth.Format("January 2006"))
	ts.calWrap.Objects = []fyne.CanvasObject{buildCalendarGrid(ts.viewingMonth, ts.dayTotals)}
	ts.calWrap.Refresh()
}

// buildCalendarGrid constructs a scrollable 7-column grid for the given month.
func buildCalendarGrid(month time.Time, dayTotals map[string]int) fyne.CanvasObject {
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	daysInMonth := first.AddDate(0, 1, 0).AddDate(0, 0, -1).Day()

	var cells []fyne.CanvasObject
	for _, h := range []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"} {
		cells = append(cells, widget.NewLabelWithStyle(h, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	}

	for i := 0; i < int(first.Weekday()); i++ {
		cells = append(cells, widget.NewLabel(""))
	}

	for d := 1; d <= daysInMonth; d++ {
		date := time.Date(month.Year(), month.Month(), d, 0, 0, 0, 0, time.Local)
		cells = append(cells, dayCell(d, dayTotals[date.Format("2006-01-02")]))
	}

	return container.NewScroll(container.NewGridWithColumns(7, cells...))
}

// dayCell creates a single calendar cell with a heatmap background.
func dayCell(day, totalSeconds int) fyne.CanvasObject {
	hours := float64(totalSeconds) / 3600

	rect := canvas.NewRectangle(heatColor(hours))
	rect.SetMinSize(fyne.NewSize(40, 40))
	rect.CornerRadius = 4

	dayLabel := widget.NewLabelWithStyle(fmt.Sprintf("%d", day), fyne.TextAlignCenter, fyne.TextStyle{})
	var hoursLabel *widget.Label
	if totalSeconds > 0 {
		hoursLabel = widget.NewLabelWithStyle(fmt.Sprintf("%.1fh", hours), fyne.TextAlignCenter, fyne.TextStyle{})
	} else {
		hoursLabel = widget.NewLabel("")
	}

	return container.NewStack(rect, container.NewVBox(dayLabel, hoursLabel))
}

// heatColor interpolates from light grey (0 h) to blue-600 (8 h+).
func heatColor(hours float64) color.Color {
	if hours <= 0 {
		return color.NRGBA{R: 240, G: 240, B: 240, A: 255}
	}
	ratio := hours / 8.0
	if ratio > 1 {
		ratio = 1
	}
	r := uint8(float64(219) - ratio*float64(219-37))
	g := uint8(float64(234) - ratio*float64(234-99))
	b := uint8(float64(254) - ratio*float64(254-235))
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}
