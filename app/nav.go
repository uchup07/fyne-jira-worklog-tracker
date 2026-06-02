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
func (a *App) buildNav() fyne.CanvasObject {
	content := container.NewStack()

	dashboard := screens.NewDashboard(a.filterState, a.worklogState, a.repo, a.fyneApp.Preferences(), a.window)
	report := screens.NewReport()         // full impl wired in Task 8
	manageTeams := screens.NewManageTeams() // full impl in Plan 4
	settings := screens.NewSettings()      // full impl in Plan 4

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
