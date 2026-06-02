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

	dashboard := screens.NewDashboard(a.filterState, a.worklogState, a.repo, a.fyneApp.Preferences(), a.window, a.tr)
	report := screens.NewReport(a.filterState, a.reportState, a.repo, a.fyneApp.Preferences(), a.window)
	manageTeams := screens.NewManageTeams(a.repo, a.tr, a.window)
	settings := screens.NewSettings(a.repo, a.fyneApp.Preferences(), a.tr, a.window, func() {
		// Rebuild nav with refreshed i18n strings when language changes
		a.window.SetContent(a.buildNav())
	})

	showScreen := func(o fyne.CanvasObject) {
		content.Objects = []fyne.CanvasObject{o}
		content.Refresh()
	}

	sidebar := container.NewVBox(
		widget.NewButtonWithIcon(a.tr.T("nav.dashboard"), theme.HomeIcon(), func() { showScreen(dashboard.Canvas()) }),
		widget.NewButtonWithIcon(a.tr.T("nav.report"), theme.DocumentIcon(), func() { showScreen(report.Canvas()) }),
		widget.NewButtonWithIcon(a.tr.T("nav.teams"), theme.GridIcon(), func() { showScreen(manageTeams.Canvas()) }),
		widget.NewButtonWithIcon(a.tr.T("nav.settings"), theme.SettingsIcon(), func() { showScreen(settings.Canvas()) }),
	)

	showScreen(dashboard.Canvas())
	return container.NewBorder(nil, nil, sidebar, nil, content)
}
