# Plan 3: Core Screens — Dashboard + Report Fully Wired

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build all custom widgets and fully implement the Dashboard and Report screens. By the end, a real Jira search runs from the filter bar, results populate the worklog table and timesheet, and the Report screen renders a mapping table with charts.

**Architecture:** Each widget is a Go struct with a `Canvas() fyne.CanvasObject` method (simpler than embedding `widget.BaseWidget` for most cases). The Setup screen gates entry to the main nav — if no config exists, the window shows Setup; on save it transitions to the main nav. The goroutine+channel+`fyne.Do` pattern from Plan 2 exercises is used throughout.

**Prerequisites:** Plans 1 and 2 complete — `db.Repository`, `jira.Client`, and all state structs exist.

**Tech Stack:** `fyne.io/fyne/v2`, `fyne/data/binding`, `fyne/widget`, `fyne/container`, `image/draw`

---

## File Map

| File | Action | Purpose |
|---|---|---|
| `widgets/progress_bar.go` | Create | Label + progress bar driven by bindings |
| `widgets/filter_bar.go` | Create | Date inputs + author/project/team selectors + Search/Cancel buttons |
| `widgets/worklog_table.go` | Create | Virtualised widget.Table bound to WorklogState.Groups |
| `widgets/timesheet.go` | Create | Calendar heatmap using canvas.Rectangle cells |
| `screens/setup.go` | Replace | Full first-run wizard |
| `screens/dashboard.go` | Replace | Full Dashboard — filter + tabs (table / diagram / timesheet) |
| `jira/mapping.go` | Create | BuildMappingReport() + aggregators (port from mapping.ts) |
| `jira/mapping_types.go` | Create | MappingReport, MappingRow, IrqGroup types |
| `jira/mapping_test.go` | Create | Tests for aggregation logic |
| `widgets/mapping_table.go` | Create | Virtualised table for MappingReport rows |
| `widgets/pie_chart.go` | Create | canvas.Raster pie chart |
| `widgets/bar_chart.go` | Create | canvas.Raster bar chart |
| `screens/report.go` | Replace | Full Report screen |
| `app/app.go` | Modify | Add setup gate in Run() |
| `app/nav.go` | Modify | Pass state + repo + client builder to screens |

---

## Task 1: Progress Bar Widget

**Files:**
- Create: `widgets/progress_bar.go`

- [ ] **Step 1: Create `widgets/progress_bar.go`**

The widget uses bindings so the background goroutine can call `.Set()` safely without `fyne.Do`.

```go
// widgets/progress_bar.go
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

// SearchProgress displays a status label above a progress bar.
// Update it from any goroutine via the Value and Text bindings.
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

// SetDone resets the bar after a search completes.
func (p *SearchProgress) SetDone() {
	p.Value.Set(1.0)
	p.Text.Set("")
}
```

- [ ] **Step 2: Add the missing `fmt` import**

The file above references `fmt.Sprintf` — add the import:

```go
import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)
```

- [ ] **Step 3: Build to verify**

```bash
go build ./widgets/...
```

---

## Task 2: Filter Bar Widget

**Files:**
- Create: `widgets/filter_bar.go`

The filter bar exposes date entry fields and Search/Cancel buttons. It takes the `FilterState` by pointer and two callbacks.

- [ ] **Step 1: Create `widgets/filter_bar.go`**

```go
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
```

- [ ] **Step 2: Build**

```bash
go build ./widgets/...
```

---

## Task 3: Worklog Table Widget

**Files:**
- Create: `widgets/worklog_table.go`

`widget.Table` is virtualised — Fyne only renders visible rows. The table binds to `WorklogState.Groups` and refreshes when the binding changes.

- [ ] **Step 1: Create `widgets/worklog_table.go`**

```go
// widgets/worklog_table.go
package widgets

import (
	"fmt"
	"time"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/uchup07/fyne-jira-worklog-tracker/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

var worklogCols = []struct {
	header string
	width  float32
}{
	{"Work Reference", 180},
	{"Total Hours", 110},
	{"Entries", 80},
	{"Authors", 200},
	{"Last Entry", 120},
}

// WorklogTable renders WorklogGroup data in a virtualised table.
type WorklogTable struct {
	groups binding.UntypedList // []jira.WorklogGroup
	canvas fyne.CanvasObject
}

// NewWorklogTable creates a table bound to ws.Groups.
func NewWorklogTable(ws *state.WorklogState) *WorklogTable {
	t := &WorklogTable{groups: ws.Groups}

	headers := make([]fyne.CanvasObject, len(worklogCols))
	for i, col := range worklogCols {
		headers[i] = widget.NewLabelWithStyle(col.header, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}
	headerRow := container.NewHBox(headers...)

	table := widget.NewTable(
		func() (int, int) {
			return t.groups.Length(), len(worklogCols)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("") // template cell
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			item, err := t.groups.GetValue(id.Row)
			if err != nil {
				return
			}
			group, ok := item.(jira.WorklogGroup)
			if !ok {
				return
			}
			label := cell.(*widget.Label)
			switch id.Col {
			case 0:
				label.SetText(group.WorkReference)
			case 1:
				label.SetText(fmt.Sprintf("%.1fh", float64(group.TotalSeconds)/3600))
			case 2:
				label.SetText(fmt.Sprintf("%d", len(group.Items)))
			case 3:
				label.SetText(uniqueAuthors(group.Items))
			case 4:
				label.SetText(lastEntryDate(group.Items))
			}
		},
	)

	// Set column widths
	for i, col := range worklogCols {
		table.SetColumnWidth(i, col.width)
	}

	// Refresh table whenever the groups binding changes
	ws.Groups.AddListener(binding.NewDataListener(func() {
		table.Refresh()
	}))

	t.canvas = container.NewBorder(headerRow, nil, nil, nil, table)
	return t
}

// Canvas returns the Fyne canvas object.
func (t *WorklogTable) Canvas() fyne.CanvasObject { return t.canvas }

// uniqueAuthors returns a comma-separated list of unique author names.
func uniqueAuthors(items []jira.WorklogItem) string {
	seen := map[string]bool{}
	var names []string
	for _, item := range items {
		if !seen[item.Author.DisplayName] {
			seen[item.Author.DisplayName] = true
			names = append(names, item.Author.DisplayName)
		}
	}
	if len(names) > 3 {
		return fmt.Sprintf("%s +%d", names[0], len(names)-1)
	}
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}

// lastEntryDate returns the most recent Started date across all items.
func lastEntryDate(items []jira.WorklogItem) string {
	var latest time.Time
	for _, item := range items {
		if item.Started.After(latest) {
			latest = item.Started
		}
	}
	if latest.IsZero() {
		return ""
	}
	return latest.Format("2006-01-02")
}
```

- [ ] **Step 2: Build**

```bash
go build ./widgets/...
```

---

## Task 4: Timesheet Widget

**Files:**
- Create: `widgets/timesheet.go`

The timesheet renders a monthly calendar heatmap. Each day cell is coloured by total hours worked.

- [ ] **Step 1: Create `widgets/timesheet.go`**

```go
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
	ws     *state.WorklogState
	canvas fyne.CanvasObject
}

// NewTimesheet creates a timesheet bound to ws.Groups.
func NewTimesheet(ws *state.WorklogState) *Timesheet {
	ts := &Timesheet{ws: ws}
	ts.canvas = ts.build()

	// Rebuild when groups change
	ws.Groups.AddListener(binding.NewDataListener(func() {
		// Replace content — Fyne containers are mutable
		newCanvas := ts.build()
		ts.canvas.(*fyne.Container).Objects = newCanvas.(*fyne.Container).Objects
		ts.canvas.Refresh()
	}))

	return ts
}

// Canvas returns the Fyne canvas object.
func (ts *Timesheet) Canvas() fyne.CanvasObject { return ts.canvas }

// build constructs the calendar grid from the current groups data.
func (ts *Timesheet) build() fyne.CanvasObject {
	// Aggregate seconds per day from all groups
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

	// Find the month to display (from WorklogState.SearchStart, or current month)
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

	// Day-of-week headers
	headers := []fyne.CanvasObject{}
	for _, day := range []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"} {
		headers = append(headers, widget.NewLabelWithStyle(day, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	}

	// Blank cells before the first day of the month
	cells := []fyne.CanvasObject{}
	startWeekday := int(first.Weekday()) // 0=Sunday
	for i := 0; i < startWeekday; i++ {
		cells = append(cells, widget.NewLabel(""))
	}

	// Day cells
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
	return container.NewBorder(title, nil, nil, nil, grid)
}

// dayCell creates a single calendar cell with a heatmap background.
func dayCell(day, totalSeconds int) fyne.CanvasObject {
	hours := float64(totalSeconds) / 3600
	bg := heatColor(hours)

	rect := canvas.NewRectangle(bg)
	rect.SetMinSize(fyne.NewSize(40, 40))
	rect.CornerRadius = 4

	label := widget.NewLabelWithStyle(fmt.Sprintf("%d", day), fyne.TextAlignCenter, fyne.TextStyle{})

	var hoursLabel *widget.Label
	if totalSeconds > 0 {
		hoursLabel = widget.NewLabelWithStyle(fmt.Sprintf("%.1fh", hours), fyne.TextAlignCenter, fyne.TextStyle{})
	} else {
		hoursLabel = widget.NewLabel("")
	}

	return container.NewStack(rect, container.NewVBox(label, hoursLabel))
}

// heatColor returns a blue heatmap colour based on hours logged.
// 0h = light grey, 8h+ = blue-600.
func heatColor(hours float64) color.Color {
	if hours <= 0 {
		return color.NRGBA{R: 240, G: 240, B: 240, A: 255}
	}
	ratio := hours / 8.0
	if ratio > 1 {
		ratio = 1
	}
	// Interpolate from light blue to blue-600
	r := uint8(219 - ratio*(219-37))
	g := uint8(234 - ratio*(234-99))
	b := uint8(254 - ratio*(254-235))
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}
```

- [ ] **Step 2: Build**

```bash
go build ./widgets/...
```

---

## Task 5: Setup Screen (Full Implementation)

**Files:**
- Replace: `screens/setup.go`

- [ ] **Step 1: Replace `screens/setup.go` with the full wizard**

```go
// screens/setup.go
package screens

import (
	"context"
	"fmt"

	"github.com/uchup07/fyne-jira-worklog-tracker/db"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Setup is the first-run configuration wizard shown when no config exists in DB.
type Setup struct {
	repo      *db.Repository
	prefs     fyne.Preferences
	window    fyne.Window
	onSaved   func() // callback — app.Run calls this to switch to main nav
	canvas    fyne.CanvasObject
}

// NewSetup creates the Setup screen.
// onSaved is called after the user successfully saves their config.
func NewSetup(repo *db.Repository, prefs fyne.Preferences, window fyne.Window, onSaved func()) *Setup {
	s := &Setup{repo: repo, prefs: prefs, window: window, onSaved: onSaved}
	s.build()
	return s
}

// Canvas returns the Fyne canvas object.
func (s *Setup) Canvas() fyne.CanvasObject { return s.canvas }

func (s *Setup) build() {
	domainEntry := widget.NewEntry()
	domainEntry.PlaceHolder = "yourorg.atlassian.net"

	emailEntry := widget.NewEntry()
	emailEntry.PlaceHolder = "you@example.com"

	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.PlaceHolder = "Your Jira API token"

	workRefEntry := widget.NewEntry()
	workRefEntry.PlaceHolder = "customfield_10001"

	statusLabel := widget.NewLabel("")

	testBtn := widget.NewButton("Test Connection", func() {
		statusLabel.SetText("Testing...")
		go func() {
			client := jira.NewClient(domainEntry.Text, emailEntry.Text, tokenEntry.Text)
			_, err := jira.FetchProjects(context.Background(), client)
			fyne.Do(func() {
				if err != nil {
					statusLabel.SetText("Connection failed: " + err.Error())
				} else {
					statusLabel.SetText("✓ Connection successful")
				}
			})
		}()
	})

	saveBtn := widget.NewButton("Save & Continue", func() {
		if domainEntry.Text == "" || emailEntry.Text == "" || tokenEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("domain, email, and API token are required"), s.window)
			return
		}
		cfg := &db.AppConfig{
			JiraDomain:     domainEntry.Text,
			Email:          emailEntry.Text,
			ApiToken:       tokenEntry.Text,
			WorkRefFieldID: workRefEntry.Text,
		}
		if err := s.repo.Config.Save(cfg); err != nil {
			dialog.ShowError(fmt.Errorf("save config: %w", err), s.window)
			return
		}
		s.prefs.SetString("api_token", cfg.ApiToken)
		if s.onSaved != nil {
			s.onSaved()
		}
	})

	form := widget.NewForm(
		widget.NewFormItem("Jira Domain", domainEntry),
		widget.NewFormItem("Email", emailEntry),
		widget.NewFormItem("API Token", tokenEntry),
		widget.NewFormItem("Work Ref Field ID", workRefEntry),
	)

	s.canvas = container.NewCenter(
		container.NewVBox(
			widget.NewLabelWithStyle("Welcome to Jira Worklog Tracker", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Enter your Jira credentials to get started."),
			form,
			container.NewHBox(testBtn, saveBtn),
			statusLabel,
		),
	)
}
```

- [ ] **Step 2: Build**

```bash
go build ./screens/...
```

---

## Task 6: Dashboard Screen (Full Implementation) + App Wiring

**Files:**
- Replace: `screens/dashboard.go`
- Modify: `app/app.go` (add setup gate in Run)
- Modify: `app/nav.go` (pass deps to screens)

- [ ] **Step 1: Replace `screens/dashboard.go`**

```go
// screens/dashboard.go
package screens

import (
	"context"
	"fmt"

	"github.com/uchup07/fyne-jira-worklog-tracker/db"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/uchup07/fyne-jira-worklog-tracker/state"
	"github.com/uchup07/fyne-jira-worklog-tracker/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Dashboard is the main worklog-viewing screen.
type Dashboard struct {
	fs       *state.FilterState
	ws       *state.WorklogState
	repo     *db.Repository
	prefs    fyne.Preferences
	window   fyne.Window
	cancelFn context.CancelFunc
	progress *widgets.SearchProgress
	canvas   fyne.CanvasObject
}

// NewDashboard creates the Dashboard screen.
func NewDashboard(fs *state.FilterState, ws *state.WorklogState, repo *db.Repository, prefs fyne.Preferences, window fyne.Window) *Dashboard {
	d := &Dashboard{fs: fs, ws: ws, repo: repo, prefs: prefs, window: window}
	d.build()
	return d
}

// Canvas returns the Fyne canvas object.
func (d *Dashboard) Canvas() fyne.CanvasObject { return d.canvas }

func (d *Dashboard) build() {
	d.progress = widgets.NewSearchProgress()

	filterBar := widgets.NewFilterBar(d.fs, d.handleSearch, d.handleCancel)

	// Tabs
	worklogTable := widgets.NewWorklogTable(d.ws)
	timesheet := widgets.NewTimesheet(d.ws)
	diagramLabel := widget.NewLabel("Diagram — coming soon")

	tabs := container.NewAppTabs(
		container.NewTabItem("Work Reference", worklogTable.Canvas()),
		container.NewTabItem("Diagram", diagramLabel),
		container.NewTabItem("Timesheet", timesheet.Canvas()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// Sync active tab to state
	tabs.OnChanged = func(t *container.TabItem) {
		switch t.Text {
		case "Work Reference":
			d.ws.ActiveTab.Set("workreference")
		case "Diagram":
			d.ws.ActiveTab.Set("diagram")
		case "Timesheet":
			d.ws.ActiveTab.Set("timesheet")
		}
	}

	top := container.NewVBox(filterBar.Canvas(), d.progress.Canvas())
	d.canvas = container.NewBorder(top, nil, nil, nil, tabs)
}

func (d *Dashboard) handleSearch() {
	// Cancel any prior in-flight search
	if d.cancelFn != nil {
		d.cancelFn()
	}
	ctx, cancel := context.WithCancel(context.Background())
	d.cancelFn = cancel

	d.progress.Reset()
	d.ws.IsLoading.Set(true)

	go func() {
		// Build Jira client from stored config
		cfg, err := d.repo.Config.Get()
		if err != nil || cfg == nil {
			fyne.Do(func() {
				d.ws.IsLoading.Set(false)
				dialog.ShowError(fmt.Errorf("no Jira configuration found — please complete Setup"), d.window)
			})
			return
		}
		cfg.ApiToken = d.prefs.String("api_token")
		client := jira.NewClient(cfg.JiraDomain, cfg.Email, cfg.ApiToken)

		// Build filters from state
		startDate, _ := d.fs.StartDate.Get()
		endDate, _ := d.fs.EndDate.Get()
		filters := jira.SearchFilters{
			StartDate: startDate,
			EndDate:   endDate,
		}

		// Collect selected author IDs
		for i := 0; i < d.fs.SelectedAuthors.Length(); i++ {
			if v, err := d.fs.SelectedAuthors.GetValue(i); err == nil {
				if id, ok := v.(string); ok {
					filters.AuthorIDs = append(filters.AuthorIDs, id)
				}
			}
		}

		progress := make(chan jira.ProgressEvent, 20)

		// Read progress events and update UI bindings (thread-safe via bindings)
		go func() {
			for ev := range progress {
				ev := ev
				switch ev.Type {
				case "searching":
					d.progress.SetSearching(ev.Pages, ev.Found)
				case "processing":
					d.progress.SetProcessing(ev.Processed, ev.Total)
				case "finalizing":
					d.progress.SetFinalizing()
				}
			}
		}()

		groups, err := jira.FetchWorklogs(ctx, client, filters, cfg.WorkRefFieldID, progress)

		fyne.Do(func() {
			d.ws.IsLoading.Set(false)
			d.progress.SetDone()
			if err != nil {
				if err != context.Canceled {
					dialog.ShowError(fmt.Errorf("search failed: %w", err), d.window)
				}
				return
			}

			// Store start/end for timesheet month display
			startDate, _ := d.fs.StartDate.Get()
			d.ws.SearchStart.Set(startDate)
			endDate, _ := d.fs.EndDate.Get()
			d.ws.SearchEnd.Set(endDate)

			// Convert []WorklogGroup to []any and set the binding
			items := make([]any, len(groups))
			for i, g := range groups {
				items[i] = g
			}
			d.ws.Groups.Set(items)
		})
	}()
}

func (d *Dashboard) handleCancel() {
	if d.cancelFn != nil {
		d.cancelFn()
	}
}
```

- [ ] **Step 2: Add setup gate to `app/app.go`**

Add this method to `App`, and update `Run()`:

```go
// In app/app.go — update Run() and add showSetup():

func (a *App) Run() {
	cfg, _ := a.repo.Config.Get()
	if cfg == nil {
		a.showSetup()
	} else {
		a.window.SetContent(a.buildNav())
	}
	a.window.ShowAndRun()
}

func (a *App) showSetup() {
	setup := screens.NewSetup(a.repo, a.fyneApp.Preferences(), a.window, func() {
		// Called after user saves config — switch to main nav
		a.window.SetContent(a.buildNav())
	})
	a.window.SetContent(setup.Canvas())
}
```

Also add the `screens` import to `app/app.go`:
```go
import (
    "github.com/uchup07/fyne-jira-worklog-tracker/custom"
    "github.com/uchup07/fyne-jira-worklog-tracker/db"
    "github.com/uchup07/fyne-jira-worklog-tracker/jira"
    "github.com/uchup07/fyne-jira-worklog-tracker/screens"
    "github.com/uchup07/fyne-jira-worklog-tracker/state"
    "fyne.io/fyne/v2"
)
```

- [ ] **Step 3: Update `app/nav.go` to pass deps to Dashboard**

```go
// app/nav.go
package app

import (
	"github.com/uchup07/fyne-jira-worklog-tracker/screens"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (a *App) buildNav() fyne.CanvasObject {
	content := container.NewStack()

	dashboard := screens.NewDashboard(a.filterState, a.worklogState, a.repo, a.fyneApp.Preferences(), a.window)
	report := screens.NewReport()         // placeholder — full impl in Task 9
	manageTeams := screens.NewManageTeams() // placeholder — full impl in Plan 4
	settings := screens.NewSettings()      // placeholder — full impl in Plan 4

	showScreen := func(o fyne.CanvasObject) {
		content.Objects = []fyne.CanvasObject{o}
		content.Refresh()
	}

	sidebar := container.NewVBox(
		widget.NewButtonWithIcon("Dashboard", theme.HomeIcon(), func() { showScreen(dashboard.Canvas()) }),
		widget.NewButtonWithIcon("Report", theme.DocumentIcon(), func() { showScreen(report.Canvas()) }),
		widget.NewButtonWithIcon("Teams", theme.GridIcon(), func() { showScreen(manageTeams.Canvas()) }),
		widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() { showScreen(settings.Canvas()) }),
	)

	showScreen(dashboard.Canvas())
	return container.NewBorder(nil, nil, sidebar, nil, content)
}
```

- [ ] **Step 4: Build and run**

```bash
go build ./... && go run .
```
Expected:
- First run (no config): Setup wizard appears.
- Fill in real Jira credentials and click **Test Connection** — should show "✓ Connection successful".
- Click **Save & Continue** — transitions to Dashboard with filter bar and empty table.
- Set a date range and click **Search** — progress bar animates, table populates with real worklog data.

- [ ] **Step 5: Commit**

```bash
git add widgets/ screens/setup.go screens/dashboard.go app/
git commit -m "feat: Dashboard screen, Setup wizard, filter bar, worklog table, timesheet"
```

---

## Task 7: Mapping Report Types + Business Logic

**Files:**
- Create: `jira/mapping_types.go`
- Create: `jira/mapping.go`
- Create: `jira/mapping_test.go`

The mapping report joins JSW (software project) issues to IRQ (request) issues by a work-reference custom field. This is the most complex port from the Next.js app (`lib/jira/mapping.ts`, 56KB).

- [ ] **Step 1: Create `jira/mapping_types.go`**

```go
// jira/mapping_types.go
package jira

import "time"

// MappingRow is one row in the mapping report — one JSW task linked to one IRQ.
type MappingRow struct {
	JSWKey        string
	JSWSummary    string
	JSWProjectKey string
	TaskName      string // parsed category from summary
	IRQKey        string
	IRQSummary    string
	IRQStatus     string
	Vertical      string
	Company       string
	Application   string
	Module        string
	Division      string
	UATEndDate    string
	TotalSeconds  int
	WorklogMonths map[string]int // "2026-01" -> seconds
}

// MappingReport is the complete output of the mapping report pipeline.
type MappingReport struct {
	Rows              []MappingRow
	TotalCRSeconds    int
	UniqueIRQCount    int
	UniqueVerticals   int
	UniqueCompanies   int
	GeneratedAt       time.Time
}

// IrqGroup aggregates MappingRows by IRQ key — used by the collapsible report view.
type IrqGroup struct {
	IRQKey      string
	IRQSummary  string
	IRQStatus   string
	Company     string
	Application string
	Module      string
	Division    string
	Vertical    string
	UATEndDate  string
	JSWTasks    []IrqJSWTask
	TotalSeconds int
}

// IrqJSWTask is a de-duplicated JSW task within an IrqGroup.
type IrqJSWTask struct {
	Key          string
	Summary      string
	ProjectKey   string
	TaskName     string
	TotalSeconds int
}

// ChartData holds labels + values for pie and bar charts.
type ChartData struct {
	Labels []string
	Values []float64
}
```

- [ ] **Step 2: Write the failing tests for aggregation logic**

```go
// jira/mapping_test.go
package jira_test

import (
	"testing"
	"time"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
)

func TestAggregateByIrqGroups(t *testing.T) {
	rows := []jira.MappingRow{
		{JSWKey: "JSW-1", IRQKey: "IRQ-1", IRQSummary: "Feature A", Vertical: "Core", Company: "Acme", TotalSeconds: 3600},
		{JSWKey: "JSW-2", IRQKey: "IRQ-1", IRQSummary: "Feature A", Vertical: "Core", Company: "Acme", TotalSeconds: 7200},
		{JSWKey: "JSW-3", IRQKey: "IRQ-2", IRQSummary: "Feature B", Vertical: "Data", Company: "Beta", TotalSeconds: 1800},
	}

	groups := jira.AggregateByIRQ(rows)

	if len(groups) != 2 {
		t.Fatalf("expected 2 IRQ groups, got %d", len(groups))
	}

	// First group (IRQ-1)
	g1 := findGroup(groups, "IRQ-1")
	if g1 == nil {
		t.Fatal("IRQ-1 group not found")
	}
	if len(g1.JSWTasks) != 2 {
		t.Errorf("IRQ-1 tasks: got %d, want 2", len(g1.JSWTasks))
	}
	if g1.TotalSeconds != 10800 {
		t.Errorf("IRQ-1 total: got %d, want 10800", g1.TotalSeconds)
	}
}

func TestAggregateByIrqDeduplicatesTasks(t *testing.T) {
	rows := []jira.MappingRow{
		{JSWKey: "JSW-1", IRQKey: "IRQ-1", TotalSeconds: 3600},
		{JSWKey: "JSW-1", IRQKey: "IRQ-1", TotalSeconds: 3600}, // duplicate
	}
	groups := jira.AggregateByIRQ(rows)
	if len(groups) != 1 || len(groups[0].JSWTasks) != 1 {
		t.Errorf("duplicate JSW task not de-duplicated: %+v", groups)
	}
}

func TestVerticalChartData(t *testing.T) {
	rows := []jira.MappingRow{
		{Vertical: "Core", TotalSeconds: 7200},
		{Vertical: "Core", TotalSeconds: 3600},
		{Vertical: "Data", TotalSeconds: 1800},
	}
	cd := jira.VerticalChartData(rows)
	if len(cd.Labels) != 2 {
		t.Fatalf("expected 2 verticals, got %d", len(cd.Labels))
	}
	// Sorted by value descending
	if cd.Labels[0] != "Core" {
		t.Errorf("expected Core first, got %q", cd.Labels[0])
	}
	if cd.Values[0] != 3.0 { // 10800 seconds = 3h
		t.Errorf("Core hours: got %f, want 3.0", cd.Values[0])
	}
}

func findGroup(groups []jira.IrqGroup, irqKey string) *jira.IrqGroup {
	for i := range groups {
		if groups[i].IRQKey == irqKey {
			return &groups[i]
		}
	}
	return nil
}

var _ = time.Now // suppress unused import if needed
```

- [ ] **Step 3: Run to confirm failure**

```bash
go test ./jira/... -run TestAggregate -v
```
Expected: compilation failure.

- [ ] **Step 4: Create `jira/mapping.go`**

```go
// jira/mapping.go
package jira

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// BuildMappingReport fetches JSW issues with a work-reference custom field,
// then fetches the corresponding IRQ issues to join metadata. Returns a full
// MappingReport.
//
// The reference implementation is lib/jira/mapping.ts in the Next.js app.
// This port implements the "worklog date range" mode only.
func BuildMappingReport(
	ctx context.Context,
	client *Client,
	filters SearchFilters,
	cfg MappingConfig,
	progress chan<- ProgressEvent,
) (*MappingReport, error) {
	defer close(progress)

	// Step 1: fetch JSW issues in the date range with a Work Reference value
	jql := fmt.Sprintf(
		`worklogDate >= %q AND worklogDate <= %q AND "%s" is not EMPTY`,
		filters.StartDate, filters.EndDate, cfg.WorkRefFieldID,
	)
	fields := []string{"summary", "project", "worklog", cfg.WorkRefFieldID}

	var allIssues []Issue
	var nextPageToken string
	page := 0
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		body := map[string]any{"jql": jql, "fields": fields, "maxResults": 50}
		if nextPageToken != "" {
			body["nextPageToken"] = nextPageToken
		}
		var res SearchResponse
		if err := client.Post(ctx, "/search/jql", body, &res); err != nil {
			return nil, fmt.Errorf("mapping search page %d: %w", page+1, err)
		}
		allIssues = append(allIssues, res.Issues...)
		page++
		progress <- ProgressEvent{Type: "searching", Pages: page, Found: len(allIssues)}
		if res.NextPageToken == "" || len(res.Issues) < 50 {
			break
		}
		nextPageToken = res.NextPageToken
	}

	// Step 2: for each JSW issue build a MappingRow (worklogs filtered by date/author)
	var rows []MappingRow
	for i, issue := range allIssues {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		progress <- ProgressEvent{Type: "processing", Processed: i + 1, Total: len(allIssues), Current: issue.Key}

		worklogs := issue.InlineWorklogs()
		if issue.WorklogTotal() > len(worklogs) {
			all, err := fetchAllWorklogs(ctx, client, issue.Key)
			if err != nil {
				return nil, fmt.Errorf("worklogs for %s: %w", issue.Key, err)
			}
			worklogs = all
		}

		totalSecs := 0
		months := map[string]int{}
		for _, wl := range worklogs {
			t, err := parseJiraTime(wl.Started)
			if err != nil {
				continue
			}
			dateStr := t.Format("2006-01-02")
			if dateStr < filters.StartDate || dateStr > filters.EndDate {
				continue
			}
			if len(filters.AuthorIDs) > 0 && !containsStr(filters.AuthorIDs, wl.Author.AccountID) {
				continue
			}
			totalSecs += wl.TimeSpentSeconds
			months[t.Format("2006-01")] += wl.TimeSpentSeconds
		}
		if totalSecs == 0 {
			continue
		}

		workRef := issue.CustomField(cfg.WorkRefFieldID)
		rows = append(rows, MappingRow{
			JSWKey:        issue.Key,
			JSWSummary:    issue.Summary(),
			JSWProjectKey: issue.ProjectKey(),
			TaskName:      parseTaskName(issue.Summary()),
			IRQKey:        workRef, // resolved to full IRQ data in step 3
			TotalSeconds:  totalSecs,
			WorklogMonths: months,
		})
	}

	progress <- ProgressEvent{Type: "finalizing"}

	// Step 3: fetch IRQ issue metadata in batches of 25 to fill in vertical/company etc.
	if err := enrichWithIRQData(ctx, client, rows, cfg); err != nil {
		return nil, err
	}

	return buildReport(rows), nil
}

// MappingConfig holds the custom field IDs needed for the mapping report.
type MappingConfig struct {
	WorkRefFieldID  string
	VerticalFieldID string
	CompanyFieldID  string
	UatEndFieldID   string
}

// enrichWithIRQData fetches IRQ issue details in batches and fills metadata into rows.
func enrichWithIRQData(ctx context.Context, client *Client, rows []MappingRow, cfg MappingConfig) error {
	// Collect unique IRQ keys
	seen := map[string]bool{}
	var keys []string
	for _, r := range rows {
		if r.IRQKey != "" && !seen[r.IRQKey] {
			seen[r.IRQKey] = true
			keys = append(keys, r.IRQKey)
		}
	}
	if len(keys) == 0 {
		return nil
	}

	// Build a map of IRQ key → Issue by fetching in batches of 25
	irqMap := map[string]Issue{}
	batchSize := 25
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}
		batch := keys[i:end]
		quoted := make([]string, len(batch))
		for j, k := range batch {
			quoted[j] = fmt.Sprintf("%q", k)
		}
		jql := fmt.Sprintf("issue in (%s)", strings.Join(quoted, ","))
		fields := []string{"summary", "status", cfg.VerticalFieldID, cfg.CompanyFieldID, cfg.UatEndFieldID}
		body := map[string]any{"jql": jql, "fields": fields, "maxResults": 50}
		var res SearchResponse
		if err := client.Post(ctx, "/search/jql", body, &res); err != nil {
			return fmt.Errorf("fetch IRQ batch: %w", err)
		}
		for _, issue := range res.Issues {
			irqMap[issue.Key] = issue
		}
	}

	// Fill rows with IRQ metadata
	for i := range rows {
		irq, ok := irqMap[rows[i].IRQKey]
		if !ok {
			continue
		}
		rows[i].IRQSummary = irq.Summary()
		if status, ok := irq.Fields["status"].(map[string]any); ok {
			rows[i].IRQStatus, _ = status["name"].(string)
		}
		rows[i].Vertical = irq.CustomField(cfg.VerticalFieldID)
		rows[i].Company = irq.CustomField(cfg.CompanyFieldID)
		rows[i].UATEndDate = irq.CustomField(cfg.UatEndFieldID)
	}
	return nil
}

// buildReport assembles a MappingReport from rows.
func buildReport(rows []MappingRow) *MappingReport {
	totalSecs := 0
	irqs := map[string]bool{}
	verticals := map[string]bool{}
	companies := map[string]bool{}
	for _, r := range rows {
		totalSecs += r.TotalSeconds
		irqs[r.IRQKey] = true
		verticals[r.Vertical] = true
		companies[r.Company] = true
	}
	return &MappingReport{
		Rows:            rows,
		TotalCRSeconds:  totalSecs,
		UniqueIRQCount:  len(irqs),
		UniqueVerticals: len(verticals),
		UniqueCompanies: len(companies),
		GeneratedAt:     time.Now(),
	}
}

// AggregateByIRQ collapses flat MappingRows into per-IRQ groups.
func AggregateByIRQ(rows []MappingRow) []IrqGroup {
	order := []string{}
	groups := map[string]*IrqGroup{}

	for _, r := range rows {
		g, ok := groups[r.IRQKey]
		if !ok {
			g = &IrqGroup{
				IRQKey:     r.IRQKey,
				IRQSummary: r.IRQSummary,
				IRQStatus:  r.IRQStatus,
				Company:    r.Company,
				Vertical:   r.Vertical,
				UATEndDate: r.UATEndDate,
			}
			groups[r.IRQKey] = g
			order = append(order, r.IRQKey)
		}
		// De-duplicate JSW tasks
		duplicate := false
		for _, t := range g.JSWTasks {
			if t.Key == r.JSWKey {
				duplicate = true
				break
			}
		}
		if !duplicate {
			g.JSWTasks = append(g.JSWTasks, IrqJSWTask{
				Key:          r.JSWKey,
				Summary:      r.JSWSummary,
				ProjectKey:   r.JSWProjectKey,
				TaskName:     r.TaskName,
				TotalSeconds: r.TotalSeconds,
			})
		}
		g.TotalSeconds += r.TotalSeconds
	}

	result := make([]IrqGroup, 0, len(order))
	for _, key := range order {
		result = append(result, *groups[key])
	}
	return result
}

// VerticalChartData returns ChartData for a vertical breakdown bar/pie chart.
// Values are in hours, sorted descending.
func VerticalChartData(rows []MappingRow) ChartData {
	totals := map[string]float64{}
	for _, r := range rows {
		if r.Vertical == "" {
			r.Vertical = "(unknown)"
		}
		totals[r.Vertical] += float64(r.TotalSeconds) / 3600
	}
	type kv struct {
		k string
		v float64
	}
	var pairs []kv
	for k, v := range totals {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].v > pairs[j].v })

	cd := ChartData{}
	for _, p := range pairs {
		cd.Labels = append(cd.Labels, p.k)
		cd.Values = append(cd.Values, p.v)
	}
	return cd
}

// parseTaskName extracts a known task category from the issue summary.
// Mirrors the ALLOWED_TASK_TYPES logic in the Next.js app.
func parseTaskName(summary string) string {
	lower := strings.ToLower(summary)
	categories := []string{"development", "testing", "bug fix", "code review", "deployment", "documentation", "analysis"}
	for _, cat := range categories {
		if strings.Contains(lower, cat) {
			return strings.Title(cat)
		}
	}
	return "Other"
}
```

- [ ] **Step 5: Run the mapping tests**

```bash
go test ./jira/... -run "TestAggregate|TestVertical" -v
```
Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add jira/mapping_types.go jira/mapping.go jira/mapping_test.go
git commit -m "feat: mapping report pipeline — IRQ aggregation, vertical chart data, enrichment"
```

---

## Task 8: Report Widgets

**Files:**
- Create: `widgets/mapping_table.go`
- Create: `widgets/pie_chart.go`
- Create: `widgets/bar_chart.go`

- [ ] **Step 1: Create `widgets/mapping_table.go`**

```go
// widgets/mapping_table.go
package widgets

import (
	"fmt"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/uchup07/fyne-jira-worklog-tracker/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

// MappingTable renders MappingReport rows in a virtualised table.
type MappingTable struct {
	rs     *state.ReportState
	canvas fyne.CanvasObject
}

var mappingCols = []string{"IRQ Key", "JSW Key", "Task Name", "Vertical", "Company", "Hours"}

// NewMappingTable creates a table bound to rs.MappingReport.
func NewMappingTable(rs *state.ReportState) *MappingTable {
	t := &MappingTable{rs: rs}

	getRows := func() []jira.MappingRow {
		val, err := rs.MappingReport.Get()
		if err != nil || val == nil {
			return nil
		}
		report, ok := val.(*jira.MappingReport)
		if !ok {
			return nil
		}
		return report.Rows
	}

	headers := make([]fyne.CanvasObject, len(mappingCols))
	for i, col := range mappingCols {
		headers[i] = widget.NewLabelWithStyle(col, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}

	table := widget.NewTable(
		func() (int, int) { return len(getRows()), len(mappingCols) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			rows := getRows()
			if id.Row >= len(rows) {
				return
			}
			row := rows[id.Row]
			label := cell.(*widget.Label)
			switch id.Col {
			case 0:
				label.SetText(row.IRQKey)
			case 1:
				label.SetText(row.JSWKey)
			case 2:
				label.SetText(row.TaskName)
			case 3:
				label.SetText(row.Vertical)
			case 4:
				label.SetText(row.Company)
			case 5:
				label.SetText(fmt.Sprintf("%.1fh", float64(row.TotalSeconds)/3600))
			}
		},
	)

	widths := []float32{100, 100, 120, 120, 120, 80}
	for i, w := range widths {
		table.SetColumnWidth(i, w)
	}

	rs.MappingReport.AddListener(binding.NewDataListener(func() {
		table.Refresh()
	}))

	headerRow := container.NewHBox(headers...)
	t.canvas = container.NewBorder(headerRow, nil, nil, nil, table)
	return t
}

// Canvas returns the Fyne canvas object.
func (t *MappingTable) Canvas() fyne.CanvasObject { return t.canvas }
```

- [ ] **Step 2: Create `widgets/pie_chart.go`**

```go
// widgets/pie_chart.go
package widgets

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// palette is a fixed set of chart colours.
var palette = []color.NRGBA{
	{37, 99, 235, 255},   // blue-600
	{234, 88, 12, 255},   // orange-600
	{22, 163, 74, 255},   // green-600
	{147, 51, 234, 255},  // purple-600
	{220, 38, 38, 255},   // red-600
	{202, 138, 4, 255},   // yellow-600
	{20, 184, 166, 255},  // teal-500
	{236, 72, 153, 255},  // pink-500
}

// PieChart renders a pie chart using canvas.Raster.
type PieChart struct {
	data   jira.ChartData
	raster *canvas.Raster
	canvas fyne.CanvasObject
}

// NewPieChart creates a pie chart for the given data.
func NewPieChart(data jira.ChartData) *PieChart {
	p := &PieChart{data: data}

	p.raster = canvas.NewRaster(func(w, h int) image.Image {
		return drawPie(p.data, w, h)
	})
	p.raster.SetMinSize(fyne.NewSize(300, 300))

	// Build legend
	legend := container.NewVBox()
	for i, label := range data.Labels {
		col := palette[i%len(palette)]
		dot := canvas.NewRectangle(col)
		dot.SetMinSize(fyne.NewSize(12, 12))
		pct := 0.0
		total := 0.0
		for _, v := range data.Values {
			total += v
		}
		if total > 0 {
			pct = data.Values[i] / total * 100
		}
		legend.Add(container.NewHBox(dot, widget.NewLabel(
			fmt.Sprintf("%s (%.1f%%)", label, pct),
		)))
	}

	p.canvas = container.NewHBox(p.raster, legend)
	return p
}

// Canvas returns the Fyne canvas object.
func (p *PieChart) Canvas() fyne.CanvasObject { return p.canvas }

// SetData updates the chart with new data and triggers a redraw.
func (p *PieChart) SetData(data jira.ChartData) {
	p.data = data
	p.raster.Refresh()
}

// drawPie renders a pie chart onto an RGBA image.
func drawPie(data jira.ChartData, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	total := 0.0
	for _, v := range data.Values {
		total += v
	}
	if total == 0 || len(data.Values) == 0 {
		return img
	}

	cx, cy := float64(w)/2, float64(h)/2
	r := math.Min(cx, cy) * 0.85
	startAngle := -math.Pi / 2 // start at 12 o'clock

	for i, val := range data.Values {
		sweep := val / total * 2 * math.Pi
		col := palette[i%len(palette)]

		// Draw filled sector using scanline approach
		steps := int(sweep / (2 * math.Pi) * 360 * 2) // ~2 points per degree
		if steps < 4 {
			steps = 4
		}
		for s := 0; s <= steps; s++ {
			angle := startAngle + sweep*float64(s)/float64(steps)
			x2 := cx + r*math.Cos(angle)
			y2 := cy + r*math.Sin(angle)
			// Draw line from centre to edge
			drawLine(img, int(cx), int(cy), int(x2), int(y2), col)
		}
		startAngle += sweep
	}
	return img
}

// drawLine draws a line between two points using Bresenham's algorithm.
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, col color.NRGBA) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	for {
		img.SetNRGBA(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
```

Add the missing `fmt` import to `pie_chart.go`:
```go
import (
    "fmt"
    "image"
    ...
)
```

- [ ] **Step 3: Create `widgets/bar_chart.go`**

```go
// widgets/bar_chart.go
package widgets

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// BarChart renders a horizontal bar chart using canvas.Raster.
type BarChart struct {
	data   jira.ChartData
	raster *canvas.Raster
	canvas fyne.CanvasObject
}

// NewBarChart creates a bar chart for the given data.
func NewBarChart(data jira.ChartData) *BarChart {
	b := &BarChart{data: data}

	b.raster = canvas.NewRaster(func(w, h int) image.Image {
		return drawBars(b.data, w, h)
	})
	b.raster.SetMinSize(fyne.NewSize(400, 250))

	// Value labels beneath the chart
	labels := container.NewVBox()
	for i, label := range data.Labels {
		if i >= len(data.Values) {
			break
		}
		labels.Add(widget.NewLabel(fmt.Sprintf("%s: %.1fh", label, data.Values[i])))
	}

	b.canvas = container.NewVBox(b.raster, labels)
	return b
}

// Canvas returns the Fyne canvas object.
func (b *BarChart) Canvas() fyne.CanvasObject { return b.canvas }

// SetData updates the chart with new data.
func (b *BarChart) SetData(data jira.ChartData) {
	b.data = data
	b.raster.Refresh()
}

// drawBars renders horizontal bars onto an RGBA image.
func drawBars(data jira.ChartData, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	n := len(data.Values)
	if n == 0 {
		return img
	}

	maxVal := 0.0
	for _, v := range data.Values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		return img
	}

	barHeight := h / (n + 1)
	if barHeight < 4 {
		barHeight = 4
	}
	padding := barHeight / 4

	for i, val := range data.Values {
		col := palette[i%len(palette)]
		barW := int(val / maxVal * float64(w-4))
		y0 := i*(barHeight+padding) + padding
		y1 := y0 + barHeight
		for y := y0; y < y1 && y < h; y++ {
			for x := 2; x < barW+2 && x < w; x++ {
				img.SetNRGBA(x, y, col)
			}
		}
	}
	return img
}
```

- [ ] **Step 4: Build**

```bash
go build ./widgets/...
```

---

## Task 9: Report Screen (Full Implementation)

**Files:**
- Replace: `screens/report.go`

- [ ] **Step 1: Replace `screens/report.go`**

```go
// screens/report.go
package screens

import (
	"context"
	"fmt"

	"github.com/uchup07/fyne-jira-worklog-tracker/db"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/uchup07/fyne-jira-worklog-tracker/state"
	"github.com/uchup07/fyne-jira-worklog-tracker/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Report shows the mapping report — table + vertical/company charts.
type Report struct {
	fs       *state.FilterState
	rs       *state.ReportState
	repo     *db.Repository
	prefs    fyne.Preferences
	window   fyne.Window
	cancelFn context.CancelFunc
	progress *widgets.SearchProgress
	canvas   fyne.CanvasObject
}

// NewReport creates the Report screen.
func NewReport(fs *state.FilterState, rs *state.ReportState, repo *db.Repository, prefs fyne.Preferences, window fyne.Window) *Report {
	r := &Report{fs: fs, rs: rs, repo: repo, prefs: prefs, window: window}
	r.build()
	return r
}

// Canvas returns the Fyne canvas object.
func (r *Report) Canvas() fyne.CanvasObject { return r.canvas }

func (r *Report) build() {
	r.progress = widgets.NewSearchProgress()

	filterBar := widgets.NewFilterBar(r.fs, r.handleSearch, r.handleCancel)
	mappingTable := widgets.NewMappingTable(r.rs)

	// Charts — rebuilt when report changes
	chartArea := container.NewStack(widget.NewLabel("Run a search to see charts"))
	r.rs.MappingReport.AddListener(binding.NewDataListener(func() {
		val, err := r.rs.MappingReport.Get()
		if err != nil || val == nil {
			return
		}
		report := val.(*jira.MappingReport)
		verticalData := jira.VerticalChartData(report.Rows)
		pieChart := widgets.NewPieChart(verticalData)
		barChart := widgets.NewBarChart(verticalData)

		chartTabs := container.NewAppTabs(
			container.NewTabItem("Vertical (Pie)", pieChart.Canvas()),
			container.NewTabItem("Vertical (Bar)", barChart.Canvas()),
		)
		chartArea.Objects = []fyne.CanvasObject{chartTabs}
		chartArea.Refresh()
	}))

	tabs := container.NewAppTabs(
		container.NewTabItem("Mapping Table", mappingTable.Canvas()),
		container.NewTabItem("Charts", chartArea),
	)

	top := container.NewVBox(filterBar.Canvas(), r.progress.Canvas())
	r.canvas = container.NewBorder(top, nil, nil, nil, tabs)
}

func (r *Report) handleSearch() {
	if r.cancelFn != nil {
		r.cancelFn()
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.cancelFn = cancel

	r.progress.Reset()
	r.rs.IsLoading.Set(true)

	go func() {
		cfg, err := r.repo.Config.Get()
		if err != nil || cfg == nil {
			fyne.Do(func() {
				r.rs.IsLoading.Set(false)
				dialog.ShowError(fmt.Errorf("no Jira configuration — complete Setup first"), r.window)
			})
			return
		}
		cfg.ApiToken = r.prefs.String("api_token")
		client := jira.NewClient(cfg.JiraDomain, cfg.Email, cfg.ApiToken)

		startDate, _ := r.fs.StartDate.Get()
		endDate, _ := r.fs.EndDate.Get()

		var authorIDs []string
		for i := 0; i < r.fs.SelectedAuthors.Length(); i++ {
			if v, err := r.fs.SelectedAuthors.GetValue(i); err == nil {
				if id, ok := v.(string); ok {
					authorIDs = append(authorIDs, id)
				}
			}
		}

		mappingCfg := jira.MappingConfig{
			WorkRefFieldID:  cfg.WorkRefFieldID,
			VerticalFieldID: cfg.VerticalFieldID,
			CompanyFieldID:  cfg.CompanyFieldID,
			UatEndFieldID:   cfg.UatEndFieldID,
		}

		progress := make(chan jira.ProgressEvent, 20)
		go func() {
			for ev := range progress {
				ev := ev
				switch ev.Type {
				case "searching":
					r.progress.SetSearching(ev.Pages, ev.Found)
				case "processing":
					r.progress.SetProcessing(ev.Processed, ev.Total)
				case "finalizing":
					r.progress.SetFinalizing()
				}
			}
		}()

		report, err := jira.BuildMappingReport(ctx, client, jira.SearchFilters{
			StartDate: startDate,
			EndDate:   endDate,
			AuthorIDs: authorIDs,
		}, mappingCfg, progress)

		fyne.Do(func() {
			r.rs.IsLoading.Set(false)
			r.progress.SetDone()
			if err != nil {
				if err != context.Canceled {
					dialog.ShowError(fmt.Errorf("report failed: %w", err), r.window)
				}
				return
			}
			r.rs.MappingReport.Set(report)
		})
	}()
}

func (r *Report) handleCancel() {
	if r.cancelFn != nil {
		r.cancelFn()
	}
}
```

- [ ] **Step 2: Update `app/nav.go` to pass deps to Report**

```go
report := screens.NewReport(a.filterState, a.reportState, a.repo, a.fyneApp.Preferences(), a.window)
```

- [ ] **Step 3: Build and run — visual checkpoint**

```bash
go build ./... && go run .
```
Expected:
- Click **Report** in the sidebar.
- Set a date range and click **Search**.
- Progress bar animates while the mapping report builds.
- Mapping Table tab populates with rows.
- Charts tab shows a pie chart and bar chart of worklog distribution by vertical.

- [ ] **Step 4: Commit and push**

```bash
git add widgets/ screens/ app/ jira/mapping*.go
git commit -m "feat: Report screen, mapping pipeline, pie/bar charts, mapping table"
git push origin main
```

---

## Phase 3 Complete ✓

**What you have:**
- All 7 custom widgets: SearchProgress, FilterBar, WorklogTable, Timesheet, MappingTable, PieChart, BarChart
- Setup wizard gates the app on first run
- Dashboard: real Jira search with live progress, goroutine cancellation, virtualised worklog table and calendar timesheet
- Report: full mapping pipeline with IRQ enrichment, virtualised table, pie/bar charts
- Teams and Settings screens remain placeholders (Plan 4)

**Next:** Plan 4 — Polish + Ship (i18n, export, Teams/Settings screens, `fyne package`)
