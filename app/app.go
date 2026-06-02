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
