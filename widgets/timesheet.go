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

// Timesheet renders a monthly heatmap of hours logged per day.
type Timesheet struct {
	ws      *state.WorklogState
	wrap    *fyne.Container // outer container so we can swap content on refresh
}

// NewTimesheet creates a timesheet bound to ws.Groups.
func NewTimesheet(ws *state.WorklogState) *Timesheet {
	ts := &Timesheet{ws: ws}
	ts.wrap = container.NewStack(ts.buildContent())

	ws.Groups.AddListener(binding.NewDataListener(func() {
		ts.wrap.Objects = []fyne.CanvasObject{ts.buildContent()}
		ts.wrap.Refresh()
	}))

	return ts
}

// Canvas returns the Fyne canvas object.
func (ts *Timesheet) Canvas() fyne.CanvasObject { return ts.wrap }

// buildContent constructs the calendar grid from the current groups data.
func (ts *Timesheet) buildContent() fyne.CanvasObject {
	dayTotals := map[string]int{}
	length := ts.ws.Groups.Length()
	for i := 0; i < length; i++ {
		val, err := ts.ws.Groups.GetValue(i)
		if err != nil {
			continue
		}
		group, ok := val.(jira.WorklogGroup)
		if !ok {
			continue
		}
		for _, item := range group.Items {
			key := item.Started.Format("2006-01-02")
			dayTotals[key] += item.TimeSpentSeconds
		}
	}

	monthStr, _ := ts.ws.SearchStart.Get()
	displayMonth := time.Now()
	if t, err := time.Parse("2006-01-02", monthStr); err == nil {
		displayMonth = t
	}

	return buildCalendar(displayMonth, dayTotals)
}

// buildCalendar constructs a 7-column grid for the given month.
func buildCalendar(month time.Time, dayTotals map[string]int) fyne.CanvasObject {
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	daysInMonth := first.AddDate(0, 1, 0).AddDate(0, 0, -1).Day()

	headers := []fyne.CanvasObject{}
	for _, day := range []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"} {
		headers = append(headers, widget.NewLabelWithStyle(day, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	}

	cells := []fyne.CanvasObject{}
	startWeekday := int(first.Weekday())
	for i := 0; i < startWeekday; i++ {
		cells = append(cells, widget.NewLabel(""))
	}

	for d := 1; d <= daysInMonth; d++ {
		date := time.Date(month.Year(), month.Month(), d, 0, 0, 0, 0, time.Local)
		key := date.Format("2006-01-02")
		secs := dayTotals[key]
		cells = append(cells, dayCell(d, secs))
	}

	title := widget.NewLabelWithStyle(
		month.Format("January 2006"),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	grid := container.NewGridWithColumns(7, append(headers, cells...)...)
	return container.NewBorder(title, nil, nil, nil, container.NewScroll(grid))
}

// dayCell creates a single calendar cell with a heatmap background.
func dayCell(day, totalSeconds int) fyne.CanvasObject {
	hours := float64(totalSeconds) / 3600
	bg := heatColor(hours)

	rect := canvas.NewRectangle(bg)
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

// heatColor interpolates from light grey (0h) to blue-600 (8h+).
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
