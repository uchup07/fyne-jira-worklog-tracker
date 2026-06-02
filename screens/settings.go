// screens/settings.go
package screens

import (
	"context"
	"fmt"

	"github.com/uchup07/fyne-jira-worklog-tracker/db"
	"github.com/uchup07/fyne-jira-worklog-tracker/i18n"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Settings shows Jira config, public holidays, and language switcher.
type Settings struct {
	repo         *db.Repository
	prefs        fyne.Preferences
	tr           *i18n.I18n
	window       fyne.Window
	onLangChange func() // called after language switch so window repaints
	canvas       fyne.CanvasObject
}

// NewSettings creates the Settings screen.
// onLangChange is called when the user switches language — typically window.Content().Refresh.
func NewSettings(repo *db.Repository, prefs fyne.Preferences, tr *i18n.I18n, window fyne.Window, onLangChange func()) *Settings {
	s := &Settings{repo: repo, prefs: prefs, tr: tr, window: window, onLangChange: onLangChange}
	s.build()
	return s
}

// Canvas returns the Fyne canvas object.
func (s *Settings) Canvas() fyne.CanvasObject { return s.canvas }

func (s *Settings) build() {
	tabs := container.NewAppTabs(
		container.NewTabItem(s.tr.T("settings.section.jira"), s.buildJiraForm()),
		container.NewTabItem(s.tr.T("settings.section.holidays"), s.buildHolidayManager()),
		container.NewTabItem(s.tr.T("settings.section.language"), s.buildLanguagePanel()),
	)
	s.canvas = tabs
}

func (s *Settings) buildJiraForm() fyne.CanvasObject {
	cfg, _ := s.repo.Config.Get()
	if cfg == nil {
		cfg = &db.AppConfig{}
	}

	domainEntry := widget.NewEntry()
	domainEntry.Text = cfg.JiraDomain
	domainEntry.PlaceHolder = "yourorg.atlassian.net"

	emailEntry := widget.NewEntry()
	emailEntry.Text = cfg.Email

	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.Text = s.prefs.String("api_token")

	workRefEntry := widget.NewEntry()
	workRefEntry.Text = cfg.WorkRefFieldID

	verticalEntry := widget.NewEntry()
	verticalEntry.Text = cfg.VerticalFieldID

	companyEntry := widget.NewEntry()
	companyEntry.Text = cfg.CompanyFieldID

	uatEndEntry := widget.NewEntry()
	uatEndEntry.Text = cfg.UatEndFieldID

	statusLabel := widget.NewLabel("")

	testBtn := widget.NewButton(s.tr.T("settings.btn.testConnection"), func() {
		statusLabel.SetText(s.tr.T("setup.status.testing"))
		go func() {
			client := jira.NewClient(domainEntry.Text, emailEntry.Text, tokenEntry.Text)
			_, err := jira.FetchProjects(context.Background(), client)
			fyne.Do(func() {
				if err != nil {
					statusLabel.SetText(s.tr.T("setup.status.failed", map[string]any{"error": err.Error()}))
				} else {
					statusLabel.SetText(s.tr.T("setup.status.success"))
				}
			})
		}()
	})

	saveBtn := widget.NewButton(s.tr.T("settings.btn.save"), func() {
		newCfg := &db.AppConfig{
			JiraDomain:      domainEntry.Text,
			Email:           emailEntry.Text,
			ApiToken:        tokenEntry.Text,
			WorkRefFieldID:  workRefEntry.Text,
			VerticalFieldID: verticalEntry.Text,
			CompanyFieldID:  companyEntry.Text,
			UatEndFieldID:   uatEndEntry.Text,
		}
		if err := s.repo.Config.Save(newCfg); err != nil {
			dialog.ShowError(fmt.Errorf("save: %w", err), s.window)
			return
		}
		s.prefs.SetString("api_token", tokenEntry.Text)
		statusLabel.SetText(s.tr.T("settings.saved"))
	})

	form := widget.NewForm(
		widget.NewFormItem(s.tr.T("settings.field.domain"), domainEntry),
		widget.NewFormItem(s.tr.T("settings.field.email"), emailEntry),
		widget.NewFormItem(s.tr.T("settings.field.token"), tokenEntry),
		widget.NewFormItem(s.tr.T("settings.field.workRefField"), workRefEntry),
		widget.NewFormItem(s.tr.T("settings.field.verticalField"), verticalEntry),
		widget.NewFormItem(s.tr.T("settings.field.companyField"), companyEntry),
		widget.NewFormItem(s.tr.T("settings.field.uatEndField"), uatEndEntry),
	)

	return container.NewVBox(form, container.NewHBox(testBtn, saveBtn), statusLabel)
}

func (s *Settings) buildHolidayManager() fyne.CanvasObject {
	var holidays []db.PublicHoliday

	holidayList := widget.NewList(
		func() int { return len(holidays) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel(""), widget.NewLabel(""))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(holidays) {
				return
			}
			h := holidays[id]
			box := item.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(h.Date)
			box.Objects[1].(*widget.Label).SetText(h.Name)
			box.Refresh()
		},
	)

	reload := func() {
		h, _ := s.repo.Holidays.List()
		holidays = h
		holidayList.Refresh()
	}
	reload()

	addBtn := widget.NewButton(s.tr.T("settings.btn.addHoliday"), func() {
		dateEntry := widget.NewEntry()
		dateEntry.PlaceHolder = "2006-01-02"
		nameEntry := widget.NewEntry()
		nameEntry.PlaceHolder = "Holiday name"
		dialog.ShowForm(s.tr.T("settings.btn.addHoliday"), "Add", "Cancel",
			[]*widget.FormItem{
				widget.NewFormItem("Date", dateEntry),
				widget.NewFormItem("Name", nameEntry),
			},
			func(ok bool) {
				if !ok || dateEntry.Text == "" || nameEntry.Text == "" {
					return
				}
				s.repo.Holidays.Add(db.PublicHoliday{Date: dateEntry.Text, Name: nameEntry.Text})
				reload()
			}, s.window)
	})

	return container.NewBorder(nil, addBtn, nil, nil, holidayList)
}

func (s *Settings) buildLanguagePanel() fyne.CanvasObject {
	current := s.tr.Lang()
	enLabel := s.tr.T("settings.lang.en")
	idLabel := s.tr.T("settings.lang.id")
	options := []string{enLabel, idLabel}

	selected := enLabel
	if current == "id" {
		selected = idLabel
	}

	radio := widget.NewRadioGroup(options, func(choice string) {
		lang := "en"
		if choice == idLabel {
			lang = "id"
		}
		s.prefs.SetString("lang", lang)
		s.tr.SetLang(lang)
		if s.onLangChange != nil {
			s.onLangChange()
		}
	})
	radio.SetSelected(selected)

	return container.NewVBox(
		widget.NewLabel(s.tr.T("settings.section.language")+":"),
		radio,
	)
}
