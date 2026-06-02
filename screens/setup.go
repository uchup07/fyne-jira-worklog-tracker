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
	repo    *db.Repository
	prefs   fyne.Preferences
	window  fyne.Window
	onSaved func() // callback — switches to main nav after save
	canvas  fyne.CanvasObject
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
