// screens/report.go
package screens

import (
	"context"
	"fmt"

	"github.com/uchup07/fyne-jira-worklog-tracker/db"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/uchup07/fyne-jira-worklog-tracker/state"
	"github.com/uchup07/fyne-jira-worklog-tracker/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Report shows the mapping report — table + vertical/company charts.
type Report struct {
	fs       *state.FilterState
	rs       *state.ReportState
	repo     *db.Repository
	prefs    fyne.Preferences
	window   fyne.Window
	cancelFn context.CancelFunc
	progress *widgets.SearchProgress
	canvas   fyne.CanvasObject
}

// NewReport creates the Report screen.
func NewReport(fs *state.FilterState, rs *state.ReportState, repo *db.Repository, prefs fyne.Preferences, window fyne.Window) *Report {
	r := &Report{fs: fs, rs: rs, repo: repo, prefs: prefs, window: window}
	r.build()
	return r
}

// Canvas returns the Fyne canvas object.
func (r *Report) Canvas() fyne.CanvasObject { return r.canvas }

func (r *Report) build() {
	r.progress = widgets.NewSearchProgress()

	filterBar := widgets.NewFilterBar(r.fs, r.handleSearch, r.handleCancel)
	mappingTable := widgets.NewMappingTable(r.rs)

	// Charts area — rebuilt when report binding changes
	chartArea := container.NewStack(widget.NewLabel("Run a search to see charts"))
	r.rs.MappingReport.AddListener(binding.NewDataListener(func() {
		val, err := r.rs.MappingReport.Get()
		if err != nil || val == nil {
			return
		}
		report := val.(*jira.MappingReport)
		verticalData := jira.VerticalChartData(report.Rows)
		pieChart := widgets.NewPieChart(verticalData)
		barChart := widgets.NewBarChart(verticalData)

		chartTabs := container.NewAppTabs(
			container.NewTabItem("Vertical (Pie)", pieChart.Canvas()),
			container.NewTabItem("Vertical (Bar)", barChart.Canvas()),
		)
		chartArea.Objects = []fyne.CanvasObject{chartTabs}
		chartArea.Refresh()
	}))

	tabs := container.NewAppTabs(
		container.NewTabItem("Mapping Table", mappingTable.Canvas()),
		container.NewTabItem("Charts", chartArea),
	)

	top := container.NewVBox(filterBar.Canvas(), r.progress.Canvas())
	r.canvas = container.NewBorder(top, nil, nil, nil, tabs)
}

func (r *Report) handleSearch() {
	if r.cancelFn != nil {
		r.cancelFn()
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.cancelFn = cancel

	r.progress.Reset()
	r.rs.IsLoading.Set(true)

	go func() {
		cfg, err := r.repo.Config.Get()
		if err != nil || cfg == nil {
			fyne.Do(func() {
				r.rs.IsLoading.Set(false)
				dialog.ShowError(fmt.Errorf("no Jira configuration — complete Setup first"), r.window)
			})
			return
		}
		cfg.ApiToken = r.prefs.String("api_token")
		client := jira.NewClient(cfg.JiraDomain, cfg.Email, cfg.ApiToken)

		startDate, _ := r.fs.StartDate.Get()
		endDate, _ := r.fs.EndDate.Get()

		var authorIDs []string
		for i := 0; i < r.fs.SelectedAuthors.Length(); i++ {
			if v, err := r.fs.SelectedAuthors.GetValue(i); err == nil {
				if id, ok := v.(string); ok {
					authorIDs = append(authorIDs, id)
				}
			}
		}

		mappingCfg := jira.MappingConfig{
			WorkRefFieldID:  cfg.WorkRefFieldID,
			VerticalFieldID: cfg.VerticalFieldID,
			CompanyFieldID:  cfg.CompanyFieldID,
			UatEndFieldID:   cfg.UatEndFieldID,
		}

		progress := make(chan jira.ProgressEvent, 20)
		go func() {
			for ev := range progress {
				ev := ev
				switch ev.Type {
				case "searching":
					r.progress.SetSearching(ev.Pages, ev.Found)
				case "processing":
					r.progress.SetProcessing(ev.Processed, ev.Total)
				case "finalizing":
					r.progress.SetFinalizing()
				}
			}
		}()

		report, err := jira.BuildMappingReport(ctx, client, jira.SearchFilters{
			StartDate: startDate,
			EndDate:   endDate,
			AuthorIDs: authorIDs,
		}, mappingCfg, progress)

		fyne.Do(func() {
			r.rs.IsLoading.Set(false)
			r.progress.SetDone()
			if err != nil {
				if err != context.Canceled {
					dialog.ShowError(fmt.Errorf("report failed: %w", err), r.window)
				}
				return
			}
			r.rs.MappingReport.Set(report)
		})
	}()
}

func (r *Report) handleCancel() {
	if r.cancelFn != nil {
		r.cancelFn()
	}
}
