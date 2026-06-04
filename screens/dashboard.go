// screens/dashboard.go
package screens

import (
	"context"
	"errors"
	"fmt"

	"github.com/uchup07/fyne-jira-worklog-tracker/db"
	"github.com/uchup07/fyne-jira-worklog-tracker/export"
	"github.com/uchup07/fyne-jira-worklog-tracker/i18n"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/uchup07/fyne-jira-worklog-tracker/state"
	"github.com/uchup07/fyne-jira-worklog-tracker/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Dashboard is the main worklog-viewing screen.
type Dashboard struct {
	fs       *state.FilterState
	ws       *state.WorklogState
	repo     *db.Repository
	prefs    fyne.Preferences
	window   fyne.Window
	tr       *i18n.I18n
	cancelFn context.CancelFunc
	progress *widgets.SearchProgress
	canvas   fyne.CanvasObject
}

// NewDashboard creates the Dashboard screen.
func NewDashboard(fs *state.FilterState, ws *state.WorklogState, repo *db.Repository, prefs fyne.Preferences, window fyne.Window, tr *i18n.I18n) *Dashboard {
	d := &Dashboard{fs: fs, ws: ws, repo: repo, prefs: prefs, window: window, tr: tr}
	d.build()
	return d
}

// Canvas returns the Fyne canvas object.
func (d *Dashboard) Canvas() fyne.CanvasObject { return d.canvas }

func (d *Dashboard) build() {
	d.progress = widgets.NewSearchProgress()

	filterBar := widgets.NewFilterBar(d.fs, d.handleSearch, d.handleCancel, d.window)

	worklogTable := widgets.NewWorklogTable(d.ws)
	timesheet := widgets.NewTimesheet(d.ws)
	diagramLabel := widget.NewLabel("Diagram — coming soon")

	tabs := container.NewAppTabs(
		container.NewTabItem(d.tr.T("dashboard.tab.worklog"), worklogTable.Canvas()),
		container.NewTabItem(d.tr.T("dashboard.tab.diagram"), diagramLabel),
		container.NewTabItem(d.tr.T("dashboard.tab.timesheet"), timesheet.Canvas()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	tabs.OnChanged = func(t *container.TabItem) {
		switch t.Text {
		case d.tr.T("dashboard.tab.worklog"):
			d.ws.ActiveTab.Set("workreference")
		case d.tr.T("dashboard.tab.diagram"):
			d.ws.ActiveTab.Set("diagram")
		case d.tr.T("dashboard.tab.timesheet"):
			d.ws.ActiveTab.Set("timesheet")
		}
	}

	exportBar := container.NewHBox(
		widget.NewButton(d.tr.T("dashboard.export.csv"), func() {
			dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
				if err != nil || uri == nil {
					return
				}
				path := uri.URI().Path()
				uri.Close()
				if err := export.WriteCSV(path, collectGroups(d.ws)); err != nil {
					dialog.ShowError(fmt.Errorf("CSV export: %w", err), d.window)
				}
			}, d.window)
		}),
		widget.NewButton(d.tr.T("dashboard.export.excel"), func() {
			dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
				if err != nil || uri == nil {
					return
				}
				path := uri.URI().Path()
				uri.Close()
				if err := export.WriteExcel(path, collectGroups(d.ws)); err != nil {
					dialog.ShowError(fmt.Errorf("Excel export: %w", err), d.window)
				}
			}, d.window)
		}),
		widget.NewButton(d.tr.T("dashboard.export.pdf"), func() {
			dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
				if err != nil || uri == nil {
					return
				}
				path := uri.URI().Path()
				uri.Close()
				start, _ := d.ws.SearchStart.Get()
				end, _ := d.ws.SearchEnd.Get()
				if err := export.WritePDF(path, collectGroups(d.ws), start, end); err != nil {
					dialog.ShowError(fmt.Errorf("PDF export: %w", err), d.window)
				}
			}, d.window)
		}),
	)

	top := container.NewVBox(filterBar.Canvas(), d.progress.Canvas(), exportBar)
	d.canvas = container.NewBorder(top, nil, nil, nil, tabs)
}

func (d *Dashboard) handleSearch() {
	if d.cancelFn != nil {
		d.cancelFn()
	}
	ctx, cancel := context.WithCancel(context.Background())
	d.cancelFn = cancel

	d.progress.Reset()
	d.ws.IsLoading.Set(true)

	go func() {
		cfg, err := d.repo.Config.Get()
		if err != nil || cfg == nil {
			fyne.Do(func() {
				d.ws.IsLoading.Set(false)
				dialog.ShowError(errors.New(d.tr.T("dashboard.error.noConfig")), d.window)
			})
			return
		}
		cfg.ApiToken = d.prefs.String("api_token")
		client := jira.NewClient(cfg.JiraDomain, cfg.Email, cfg.ApiToken)

		startDate, _ := d.fs.StartDate.Get()
		endDate, _ := d.fs.EndDate.Get()
		filters := jira.SearchFilters{StartDate: startDate, EndDate: endDate}

		for i := 0; i < d.fs.SelectedAuthors.Length(); i++ {
			if v, err := d.fs.SelectedAuthors.GetValue(i); err == nil {
				if id, ok := v.(string); ok {
					filters.AuthorIDs = append(filters.AuthorIDs, id)
				}
			}
		}

		progress := make(chan jira.ProgressEvent, 20)
		go func() {
			for ev := range progress {
				ev := ev
				switch ev.Type {
				case "searching":
					d.progress.SetSearching(ev.Pages, ev.Found)
				case "processing":
					d.progress.SetProcessing(ev.Processed, ev.Total)
				case "finalizing":
					d.progress.SetFinalizing()
				}
			}
		}()

		groups, err := jira.FetchWorklogs(ctx, client, filters, cfg.WorkRefFieldID, progress)

		fyne.Do(func() {
			d.ws.IsLoading.Set(false)
			d.progress.SetDone()
			if err != nil {
				if err != context.Canceled {
					dialog.ShowError(errors.New(d.tr.T("dashboard.error.searchFailed", map[string]any{"error": err.Error()})), d.window)
				}
				return
			}
			startDate, _ := d.fs.StartDate.Get()
			d.ws.SearchStart.Set(startDate)
			endDate, _ := d.fs.EndDate.Get()
			d.ws.SearchEnd.Set(endDate)

			items := make([]any, len(groups))
			for i, g := range groups {
				items[i] = g
			}
			d.ws.Groups.Set(items)
		})
	}()
}

func (d *Dashboard) handleCancel() {
	if d.cancelFn != nil {
		d.cancelFn()
	}
}

// collectGroups extracts all WorklogGroup values from the state binding.
func collectGroups(ws *state.WorklogState) []jira.WorklogGroup {
	var groups []jira.WorklogGroup
	for i := 0; i < ws.Groups.Length(); i++ {
		val, err := ws.Groups.GetValue(i)
		if err != nil {
			continue
		}
		if g, ok := val.(jira.WorklogGroup); ok {
			groups = append(groups, g)
		}
	}
	return groups
}
