// app/app.go
package app

import (
	"github.com/uchup07/fyne-jira-worklog-tracker/custom"
	"github.com/uchup07/fyne-jira-worklog-tracker/db"
	"github.com/uchup07/fyne-jira-worklog-tracker/i18n"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/uchup07/fyne-jira-worklog-tracker/screens"
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
	tr   *i18n.I18n
}

// New initialises the App. DB is opened from Fyne's app-storage directory.
func New(a fyne.App, w fyne.Window) *App {
	a.Settings().SetTheme(custom.NewAppTheme())

	dbPath := a.Storage().RootURI().Path() + "/worklog.db"
	repo := db.Open(dbPath)

	lang := a.Preferences().StringWithFallback("lang", "en")

	return &App{
		fyneApp:      a,
		window:       w,
		filterState:  state.NewFilterState(),
		worklogState: state.NewWorklogState(),
		reportState:  state.NewReportState(),
		repo:         repo,
		tr:           i18n.New(lang),
	}
}

// Run checks for saved config: shows Setup wizard on first run, else shows main nav.
func (a *App) Run() {
	cfg, _ := a.repo.Config.Get()
	if cfg == nil {
		a.showSetup()
	} else {
		a.window.SetContent(a.buildNav())
	}
	a.window.ShowAndRun()
}

// showSetup presents the first-run wizard. On save it transitions to the main nav.
func (a *App) showSetup() {
	setup := screens.NewSetup(a.repo, a.fyneApp.Preferences(), a.window, func() {
		a.window.SetContent(a.buildNav())
	})
	a.window.SetContent(setup.Canvas())
}

// JiraClient builds a Jira client from stored config + OS-keychain token.
// Returns nil if config is missing or token is empty.
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
