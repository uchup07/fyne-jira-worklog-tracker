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
