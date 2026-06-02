// app/app.go
package app

import (
	"github.com/uchup07/fyne-jira-worklog-tracker/custom"
	"github.com/uchup07/fyne-jira-worklog-tracker/db"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/uchup07/fyne-jira-worklog-tracker/state"

	"fyne.io/fyne/v2"
)

// App is the root of the application.
type App struct {
	fyneApp fyne.App
	window  fyne.Window

	filterState  *state.FilterState
	worklogState *state.WorklogState
	reportState  *state.ReportState

	repo *db.Repository
}

// New initialises the App. DB is opened from Fyne's app-storage directory
// so it persists between runs on the user's machine.
func New(a fyne.App, w fyne.Window) *App {
	a.Settings().SetTheme(custom.NewAppTheme())

	dbPath := a.Storage().RootURI().Path() + "/worklog.db"
	repo := db.Open(dbPath)

	return &App{
		fyneApp:      a,
		window:       w,
		filterState:  state.NewFilterState(),
		worklogState: state.NewWorklogState(),
		reportState:  state.NewReportState(),
		repo:         repo,
	}
}

// Run wires navigation and starts the Fyne event loop.
func (a *App) Run() {
	a.window.SetContent(a.buildNav())
	a.window.ShowAndRun()
}

// JiraClient builds a Jira client from the stored config + the OS-keychain token.
// Returns nil if config is not yet saved or token is missing.
func (a *App) JiraClient() *jira.Client {
	cfg, err := a.repo.Config.Get()
	if err != nil || cfg == nil {
		return nil
	}
	cfg.ApiToken = a.fyneApp.Preferences().String("api_token")
	if cfg.ApiToken == "" {
		return nil
	}
	return jira.NewClient(cfg.JiraDomain, cfg.Email, cfg.ApiToken)
}
