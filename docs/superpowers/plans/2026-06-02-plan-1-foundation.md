# Plan 1: Foundation — Go Fundamentals + Fyne Skeleton + State Layer

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Set up the project, learn Go fundamentals through targeted exercises, build the Fyne app window with sidebar navigation and 5 placeholder screens, and implement the reactive state layer.

**Architecture:** No business logic yet. This plan produces a working app window where the sidebar switches between placeholder screens, and a state layer whose bindings auto-refresh bound widgets. All subsequent plans build on this foundation.

**Tech Stack:** Go 1.23, Fyne.io v2.5+, fyne/data/binding

---

## File Map

| File | Action | Purpose |
|---|---|---|
| `learn/01_structs/main.go` | Create | Exercise: structs, methods, error handling |
| `learn/02_concurrency/main.go` | Create | Exercise: goroutines, channels, context |
| `go.mod` | Create | Module declaration + dependencies |
| `.gitignore` | Create | Ignore build artifacts |
| `assets/icon.png` | Create | Placeholder 1×1 PNG (replaced later) |
| `main.go` | Create | App entry point |
| `app/app.go` | Create | App struct — window + state |
| `app/nav.go` | Create | Sidebar navigation builder |
| `custom/theme.go` | Create | AppTheme — blue-600 primary colour |
| `screens/dashboard.go` | Create | Placeholder Dashboard screen |
| `screens/report.go` | Create | Placeholder Report screen |
| `screens/manage_teams.go` | Create | Placeholder Manage Teams screen |
| `screens/settings.go` | Create | Placeholder Settings screen |
| `screens/setup.go` | Create | Placeholder Setup screen |
| `state/filter.go` | Create | FilterState with binding.* fields |
| `state/worklog.go` | Create | WorklogState with binding.* fields |
| `state/report.go` | Create | ReportState with binding.* fields |
| `state/state_test.go` | Create | Tests for default values and listener behaviour |

---

## Task 1: Verify Prerequisites

**Files:** none (shell commands only)

- [ ] **Step 1: Check Go is installed**

```bash
go version
```
Expected: `go version go1.23.x ...` (1.21+ is fine). If missing, install from https://go.dev/dl/

- [ ] **Step 2: Check Xcode CLI tools (required for Fyne's OpenGL/CGO)**

```bash
xcode-select --print-path
```
Expected: `/Applications/Xcode.app/Contents/Developer` or `/Library/Developer/CommandLineTools`

If the command fails:
```bash
xcode-select --install
```

- [ ] **Step 3: Check git is available**

```bash
git --version
```
Expected: `git version 2.x.x`

- [ ] **Step 4: Confirm working directory is the project root**

```bash
pwd
ls docs/superpowers/specs/
```
Expected: you're in `jira-worklog-tracker/` and the spec file is visible.

---

## Task 2: Go Fundamentals Exercise — Structs, Methods, Interfaces

**Files:**
- Create: `learn/01_structs/main.go`

These patterns are used directly in the app. Complete the exercise before writing any app code.

- [ ] **Step 1: Create the exercise file**

```go
// learn/01_structs/main.go
package main

import (
	"fmt"
	"time"
)

// --- Structs and methods ---

// WorklogItem represents a single worklog entry (mirrors the real app type).
type WorklogItem struct {
	IssueKey         string
	Author           string
	TimeSpentSeconds int
	Started          time.Time
}

// Hours returns the time spent as fractional hours.
func (w WorklogItem) Hours() float64 {
	return float64(w.TimeSpentSeconds) / 3600
}

// WorklogGroup groups items by work reference (mirrors the real app type).
type WorklogGroup struct {
	WorkReference string
	Items         []WorklogItem
}

// TotalHours sums hours across all items in the group.
func (g WorklogGroup) TotalHours() float64 {
	total := 0.0
	for _, item := range g.Items {
		total += item.Hours()
	}
	return total
}

// --- Interfaces ---

// Summariser is any type that can produce a one-line summary string.
// In the real app, fyne.Widget is a similar interface — anything that
// implements the right methods can be used as a UI element.
type Summariser interface {
	Summary() string
}

func (g WorklogGroup) Summary() string {
	return fmt.Sprintf("%s: %.1fh (%d entries)", g.WorkReference, g.TotalHours(), len(g.Items))
}

// printAll prints the summary of anything that implements Summariser.
func printAll(items []Summariser) {
	for _, item := range items {
		fmt.Println(item.Summary())
	}
}

// --- Error handling ---

// parseHours converts a "Xh Ym" string to total seconds.
// Go functions return errors explicitly — no exceptions.
func parseHours(s string) (int, error) {
	var h, m int
	_, err := fmt.Sscanf(s, "%dh %dm", &h, &m)
	if err != nil {
		return 0, fmt.Errorf("parseHours: invalid format %q (expected e.g. '2h 30m'): %w", s, err)
	}
	return h*3600 + m*60, nil
}

func main() {
	// Build two groups
	groups := []WorklogGroup{
		{
			WorkReference: "CR-001",
			Items: []WorklogItem{
				{IssueKey: "JSW-1", Author: "Alice", TimeSpentSeconds: 7200, Started: time.Now()},
				{IssueKey: "JSW-2", Author: "Bob", TimeSpentSeconds: 3600, Started: time.Now()},
			},
		},
		{
			WorkReference: "CR-002",
			Items: []WorklogItem{
				{IssueKey: "JSW-3", Author: "Alice", TimeSpentSeconds: 5400, Started: time.Now()},
			},
		},
	}

	// Use interface — WorklogGroup satisfies Summariser
	summaries := make([]Summariser, len(groups))
	for i, g := range groups {
		summaries[i] = g
	}
	printAll(summaries)

	// Error handling
	secs, err := parseHours("2h 30m")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("2h 30m = %d seconds\n", secs)
	}

	_, err = parseHours("bad input")
	if err != nil {
		fmt.Println("Expected error:", err)
	}
}
```

- [ ] **Step 2: Run the exercise**

```bash
go run learn/01_structs/main.go
```
Expected output:
```
CR-001: 3.0h (2 entries)
CR-002: 1.5h (1 entries)
2h 30m = 9000 seconds
Expected error: parseHours: invalid format "bad input" ...
```

---

## Task 3: Go Fundamentals Exercise — Goroutines, Channels, Context

**Files:**
- Create: `learn/02_concurrency/main.go`

This is the most important exercise — the goroutine+channel pattern runs throughout the entire app.

- [ ] **Step 1: Create the exercise file**

```go
// learn/02_concurrency/main.go
package main

import (
	"context"
	"fmt"
	"time"
)

// ProgressEvent mirrors the real app's type.
type ProgressEvent struct {
	Type      string // "searching" | "processing" | "done"
	Processed int
	Total     int
}

// simulateSearch mimics a Jira worklog fetch.
// Key ideas:
//   - Runs on its own goroutine (caller does: go simulateSearch(...))
//   - Sends progress updates via a channel (non-blocking, buffered)
//   - Respects ctx.Done() to cancel early
//   - Closes the channel when done (signals the reader to stop)
func simulateSearch(ctx context.Context, total int, progress chan<- ProgressEvent) {
	defer close(progress) // ALWAYS close when done — reader loop ends on close

	for i := 1; i <= total; i++ {
		// Check for cancellation before each unit of work
		select {
		case <-ctx.Done():
			fmt.Println("worker: cancelled at item", i)
			return
		default:
		}

		// Simulate doing work
		time.Sleep(100 * time.Millisecond)

		// Send progress — won't block because channel is buffered
		progress <- ProgressEvent{
			Type:      "processing",
			Processed: i,
			Total:     total,
		}
	}
	progress <- ProgressEvent{Type: "done"}
}

func main() {
	fmt.Println("--- Example 1: Normal completion ---")
	runSearch(context.Background(), 5)

	fmt.Println("\n--- Example 2: Cancellation after 2 items ---")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(250 * time.Millisecond) // cancel after ~2 items
		cancel()
	}()
	runSearch(ctx, 10)
}

func runSearch(ctx context.Context, total int) {
	progress := make(chan ProgressEvent, 10) // buffered — worker never blocks

	go simulateSearch(ctx, total, progress)

	// Read all events until the channel is closed.
	// "range ch" exits automatically when ch is closed.
	for ev := range progress {
		switch ev.Type {
		case "processing":
			fmt.Printf("  processed %d / %d\n", ev.Processed, ev.Total)
		case "done":
			fmt.Println("  done!")
		}
	}
	// When we get here, the channel is closed — worker has finished or was cancelled.
	fmt.Println("search complete")
}
```

- [ ] **Step 2: Run the exercise**

```bash
go run learn/02_concurrency/main.go
```
Expected output (cancel timing is approximate):
```
--- Example 1: Normal completion ---
  processed 1 / 5
  processed 2 / 5
  processed 3 / 5
  processed 4 / 5
  processed 5 / 5
  done!
search complete

--- Example 2: Cancellation after 2 items ---
  processed 1 / 10
  processed 2 / 10
worker: cancelled at item 3
search complete
```

---

## Task 4: Project Scaffold

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `assets/icon.png` (1×1 placeholder)

- [ ] **Step 1: Initialise the Go module**

```bash
go mod init github.com/uchup07/fyne-jira-worklog-tracker
```

- [ ] **Step 2: Install dependencies**

```bash
go get fyne.io/fyne/v2@latest
go get modernc.org/sqlite@latest
go get github.com/xuri/excelize/v2@latest
go get github.com/jung-kurt/gofpdf@latest
go get github.com/nicksnyder/go-i18n/v2@latest
go get golang.org/x/text@latest
go mod tidy
```

Expected: `go.sum` is created; no error output.

> **Note:** `fyne.io/fyne/v2` requires CGO for its OpenGL renderer. On macOS, Xcode CLI tools (Task 1) provide the C compiler automatically.

- [ ] **Step 3: Create the directory structure**

```bash
mkdir -p app screens widgets state jira db export i18n/locales custom assets learn/01_structs learn/02_concurrency
```

- [ ] **Step 4: Create `.gitignore`**

```
# Go build output
*.exe
*.app
*.dmg

# Fyne bundle output
FyneApp.toml

# OS files
.DS_Store

# SQLite databases (local state only)
*.db

# Test cache
.testcache/
```

- [ ] **Step 5: Create a 1×1 placeholder icon**

```bash
# Create a minimal valid PNG (1×1 blue pixel) using Go
cat > assets/gen_icon.go << 'EOF'
//go:build ignore

package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{37, 99, 235, 255})
	f, _ := os.Create("icon.png")
	defer f.Close()
	png.Encode(f, img)
}
EOF
cd assets && go run gen_icon.go && rm gen_icon.go && cd ..
```

Expected: `assets/icon.png` now exists.

- [ ] **Step 6: Commit scaffold**

```bash
git add go.mod go.sum .gitignore assets/
git commit -m "feat: project scaffold — go.mod, deps, assets"
```

---

## Task 5: App Entry Point + App Struct + Theme

**Files:**
- Create: `main.go`
- Create: `app/app.go`
- Create: `custom/theme.go`

- [ ] **Step 1: Create `custom/theme.go`**

```go
// custom/theme.go
package custom

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// AppTheme overrides the default Fyne theme with the app's brand colours.
type AppTheme struct {
	fyne.Theme
}

// NewAppTheme returns the configured app theme.
func NewAppTheme() fyne.Theme {
	return &AppTheme{Theme: theme.DefaultTheme()}
}

// Color overrides selected colour tokens. All others fall back to DefaultTheme.
func (t *AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 37, G: 99, B: 235, A: 255} // blue-600
	case theme.ColorNameFocus:
		return color.NRGBA{R: 37, G: 99, B: 235, A: 180}
	}
	return t.Theme.Color(name, variant)
}
```

- [ ] **Step 2: Create `app/app.go`** (skeleton — DB and Jira client are added in Plan 2)

```go
// app/app.go
package app

import (
	"github.com/uchup07/fyne-jira-worklog-tracker/custom"
	"github.com/uchup07/fyne-jira-worklog-tracker/state"

	"fyne.io/fyne/v2"
)

// App is the root of the application. It owns the window and all shared state.
// Services (DB, Jira client) are added in Plan 2.
type App struct {
	fyneApp fyne.App
	window  fyne.Window

	filterState  *state.FilterState
	worklogState *state.WorklogState
	reportState  *state.ReportState
}

// New creates the App. Call Run() after New().
func New(a fyne.App, w fyne.Window) *App {
	a.Settings().SetTheme(custom.NewAppTheme())
	return &App{
		fyneApp:      a,
		window:       w,
		filterState:  state.NewFilterState(),
		worklogState: state.NewWorklogState(),
		reportState:  state.NewReportState(),
	}
}

// Run wires navigation and starts the Fyne event loop (blocks until window closes).
func (a *App) Run() {
	a.window.SetContent(a.buildNav())
	a.window.ShowAndRun()
}
```

- [ ] **Step 3: Create `main.go`**

```go
// main.go
package main

import (
	"github.com/uchup07/fyne-jira-worklog-tracker/app"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
)

func main() {
	a := fyneapp.NewWithID("com.uchup07.jira-worklog-tracker")
	w := a.NewWindow("Jira Worklog Tracker")
	w.Resize(fyne.NewSize(1280, 800))
	w.SetMaster()

	tracker := app.New(a, w)
	tracker.Run()
}
```

- [ ] **Step 4: Verify it compiles (will fail at runtime — nav not built yet)**

```bash
go build ./...
```
Expected: compilation errors about missing `buildNav` and screens — that's fine; we add them next.

---

## Task 6: Navigation + Placeholder Screens

**Files:**
- Create: `screens/dashboard.go`
- Create: `screens/report.go`
- Create: `screens/manage_teams.go`
- Create: `screens/settings.go`
- Create: `screens/setup.go`
- Create: `app/nav.go`

- [ ] **Step 1: Create the 5 placeholder screen files**

```go
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
```

```go
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
```

```go
// screens/manage_teams.go
package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// ManageTeams is the team/member CRUD screen.
// Full implementation comes in Plan 4.
type ManageTeams struct {
	canvas fyne.CanvasObject
}

func NewManageTeams() *ManageTeams {
	m := &ManageTeams{}
	m.canvas = widget.NewLabel("Manage Teams — full implementation in Plan 4")
	return m
}

func (m *ManageTeams) Canvas() fyne.CanvasObject { return m.canvas }
```

```go
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
```

```go
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
```

- [ ] **Step 2: Create `app/nav.go`**

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

// buildNav constructs the sidebar + content area layout.
// The sidebar has 4 buttons; each swaps the content area to a different screen.
func (a *App) buildNav() fyne.CanvasObject {
	// content is a Stack — shows exactly one screen at a time
	content := container.NewStack()

	// Create all screens (lazy — no data yet)
	dashboard := screens.NewDashboard()
	report := screens.NewReport()
	manageTeams := screens.NewManageTeams()
	settings := screens.NewSettings()

	// showScreen replaces the content stack with the given canvas object
	showScreen := func(o fyne.CanvasObject) {
		content.Objects = []fyne.CanvasObject{o}
		content.Refresh()
	}

	// Build sidebar buttons
	sidebar := container.NewVBox(
		widget.NewButtonWithIcon("Dashboard", theme.HomeIcon(),
			func() { showScreen(dashboard.Canvas()) }),
		widget.NewButtonWithIcon("Report", theme.DocumentIcon(),
			func() { showScreen(report.Canvas()) }),
		widget.NewButtonWithIcon("Teams", theme.GridIcon(),
			func() { showScreen(manageTeams.Canvas()) }),
		widget.NewButtonWithIcon("Settings", theme.SettingsIcon(),
			func() { showScreen(settings.Canvas()) }),
	)

	// Show Dashboard by default
	showScreen(dashboard.Canvas())

	// Border layout: sidebar on the left, content fills the rest
	return container.NewBorder(nil, nil, sidebar, nil, content)
}
```

- [ ] **Step 3: Build and run**

```bash
go build ./... && go run .
```
Expected: app window opens (1280×800), sidebar shows 4 buttons, clicking each swaps the label in the main area.

- [ ] **Step 4: Commit**

```bash
git add app/ screens/ custom/ main.go
git commit -m "feat: Fyne app skeleton — window, sidebar nav, placeholder screens, theme"
```

---

## Task 7: State Layer

**Files:**
- Create: `state/filter.go`
- Create: `state/worklog.go`
- Create: `state/report.go`

The state layer uses `fyne/data/binding`. A binding is a reactive value: when you call `.Set()`, any widget or listener bound to it is automatically notified. This replaces Zustand from the Next.js app.

- [ ] **Step 1: Create `state/filter.go`**

```go
// state/filter.go
package state

import (
	"time"

	"fyne.io/fyne/v2/data/binding"
)

// FilterState holds all user-selected filter criteria for a worklog search.
// Fields are binding types so bound widgets auto-refresh when values change.
type FilterState struct {
	StartDate        binding.String      // "2006-01-02"
	EndDate          binding.String      // "2006-01-02"
	SelectedAuthors  binding.UntypedList // []string — Jira accountId values
	SelectedProjects binding.UntypedList // []string — Jira project keys
	SelectedTeamIDs  binding.UntypedList // []int — internal DB team IDs
	DateMode         binding.String      // "day" | "week" | "month" | "between"
}

// NewFilterState creates FilterState with sensible defaults (today's date, day mode).
func NewFilterState() *FilterState {
	today := time.Now().Format("2006-01-02")
	s := &FilterState{
		StartDate:        binding.NewString(),
		EndDate:          binding.NewString(),
		SelectedAuthors:  binding.NewUntypedList(),
		SelectedProjects: binding.NewUntypedList(),
		SelectedTeamIDs:  binding.NewUntypedList(),
		DateMode:         binding.NewString(),
	}
	_ = s.StartDate.Set(today)
	_ = s.EndDate.Set(today)
	_ = s.DateMode.Set("day")
	return s
}
```

- [ ] **Step 2: Create `state/worklog.go`**

```go
// state/worklog.go
package state

import "fyne.io/fyne/v2/data/binding"

// WorklogState holds the results and loading state for the Dashboard screen.
type WorklogState struct {
	// Groups holds []jira.WorklogGroup — stored as UntypedList so any widget
	// can listen for changes without importing the jira package directly.
	Groups      binding.UntypedList
	SearchStart binding.String // date of the last completed search
	SearchEnd   binding.String
	ActiveTab   binding.String // "workreference" | "diagram" | "timesheet"
	IsLoading   binding.Bool
}

// NewWorklogState creates WorklogState with default tab selection.
func NewWorklogState() *WorklogState {
	s := &WorklogState{
		Groups:      binding.NewUntypedList(),
		SearchStart: binding.NewString(),
		SearchEnd:   binding.NewString(),
		ActiveTab:   binding.NewString(),
		IsLoading:   binding.NewBool(),
	}
	_ = s.ActiveTab.Set("workreference")
	return s
}
```

- [ ] **Step 3: Create `state/report.go`**

```go
// state/report.go
package state

import "fyne.io/fyne/v2/data/binding"

// ReportState holds the results and loading state for the Report screen.
type ReportState struct {
	// MappingReport holds *jira.MappingReport — stored as Untyped to avoid
	// a circular import between state and jira packages.
	MappingReport binding.Untyped
	IsLoading     binding.Bool
}

// NewReportState creates ReportState with zero values.
func NewReportState() *ReportState {
	return &ReportState{
		MappingReport: binding.NewUntyped(),
		IsLoading:     binding.NewBool(),
	}
}
```

- [ ] **Step 4: Build to verify no errors**

```bash
go build ./...
```
Expected: clean build (no output).

---

## Task 8: State Tests

**Files:**
- Create: `state/state_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// state/state_test.go
package state_test

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/data/binding"

	"github.com/uchup07/fyne-jira-worklog-tracker/state"
)

func TestFilterStateDefaults(t *testing.T) {
	fs := state.NewFilterState()
	today := time.Now().Format("2006-01-02")

	start, err := fs.StartDate.Get()
	if err != nil {
		t.Fatalf("StartDate.Get: %v", err)
	}
	if start != today {
		t.Errorf("StartDate default: got %q, want %q", start, today)
	}

	end, _ := fs.EndDate.Get()
	if end != today {
		t.Errorf("EndDate default: got %q, want %q", end, today)
	}

	mode, _ := fs.DateMode.Get()
	if mode != "day" {
		t.Errorf("DateMode default: got %q, want %q", mode, "day")
	}
}

func TestFilterStateSet(t *testing.T) {
	fs := state.NewFilterState()

	if err := fs.StartDate.Set("2026-01-01"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, _ := fs.StartDate.Get()
	if got != "2026-01-01" {
		t.Errorf("got %q, want %q", got, "2026-01-01")
	}
}

func TestWorklogStateDefaults(t *testing.T) {
	ws := state.NewWorklogState()

	tab, err := ws.ActiveTab.Get()
	if err != nil {
		t.Fatalf("ActiveTab.Get: %v", err)
	}
	if tab != "workreference" {
		t.Errorf("ActiveTab default: got %q, want %q", tab, "workreference")
	}

	loading, _ := ws.IsLoading.Get()
	if loading {
		t.Error("IsLoading should default to false")
	}
}

func TestBindingListenerFires(t *testing.T) {
	ws := state.NewWorklogState()

	fired := make(chan struct{}, 1)
	ws.IsLoading.AddListener(binding.NewDataListener(func() {
		// Non-blocking send — if channel already has a value, don't block
		select {
		case fired <- struct{}{}:
		default:
		}
	}))

	if err := ws.IsLoading.Set(true); err != nil {
		t.Fatalf("Set: %v", err)
	}

	select {
	case <-fired:
		// pass — listener was notified
	default:
		t.Error("listener was not notified after Set(true)")
	}
}

func TestReportStateDefaults(t *testing.T) {
	rs := state.NewReportState()

	val, err := rs.MappingReport.Get()
	if err != nil {
		t.Fatalf("MappingReport.Get: %v", err)
	}
	if val != nil {
		t.Error("MappingReport should default to nil")
	}
}
```

- [ ] **Step 2: Run the failing tests**

```bash
go test ./state/... -v
```
Expected: PASS for all tests (the state package is pure Go — no Fyne display needed).

---

## Task 9: Phase 1 Checkpoint + Push

- [ ] **Step 1: Full build**

```bash
go build ./...
```
Expected: no errors.

- [ ] **Step 2: Run all tests**

```bash
go test ./...
```
Expected: all tests pass, output similar to:
```
ok  	github.com/uchup07/fyne-jira-worklog-tracker/state	0.003s
```

- [ ] **Step 3: Visual checkpoint — run the app**

```bash
go run .
```
Expected: 1280×800 window opens with a blue-tinted sidebar and 4 nav buttons. Each button swaps the placeholder label in the main area. Close the window to exit.

- [ ] **Step 4: Commit and push**

```bash
git add learn/ state/ go.mod go.sum .gitignore
git commit -m "feat: state layer with binding types + tests; Go fundamentals exercises"
git push origin main
```

---

## Phase 1 Complete ✓

**What you have:**
- Working Fyne app window with sidebar navigation
- 5 placeholder screens (Dashboard, Report, Teams, Settings, Setup)
- Reactive state layer (`FilterState`, `WorklogState`, `ReportState`) with tests
- Blue-600 theme applied globally
- Go fundamentals exercises covering structs, interfaces, goroutines, channels, and context

**Next:** Plan 2 — Data Services (Jira HTTP client + SQLite DB layer)
