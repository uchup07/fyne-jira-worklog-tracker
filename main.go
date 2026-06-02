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
