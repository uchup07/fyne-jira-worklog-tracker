// app/nav.go
package app

import (
	"github.com/uchup07/fyne-jira-worklog-tracker/screens"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// sidebarLayout is a two-column layout with a fixed-width left panel.
type sidebarLayout struct{ width float32 }

func (l *sidebarLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(l.width, size.Height))
	objects[1].Move(fyne.NewPos(l.width, 0))
	objects[1].Resize(fyne.NewSize(size.Width-l.width, size.Height))
}

func (l *sidebarLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minH := float32(0)
	for _, o := range objects {
		if h := o.MinSize().Height; h > minH {
			minH = h
		}
	}
	return fyne.NewSize(l.width, minH)
}

// themedSidebarBG renders a rounded rectangle whose fill color tracks the active
// Fyne theme, so it adapts correctly when the OS switches dark/light mode.
type themedSidebarBG struct {
	widget.BaseWidget
}

func newThemedSidebarBG() *themedSidebarBG {
	b := &themedSidebarBG{}
	b.ExtendBaseWidget(b)
	return b
}

func (b *themedSidebarBG) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(theme.OverlayBackgroundColor())
	rect.CornerRadius = 10
	return &sidebarBGRenderer{rect: rect}
}

type sidebarBGRenderer struct {
	rect *canvas.Rectangle
}

func (r *sidebarBGRenderer) Layout(s fyne.Size) {
	r.rect.Move(fyne.NewPos(0, 0))
	r.rect.Resize(s)
}

func (r *sidebarBGRenderer) MinSize() fyne.Size           { return fyne.NewSize(0, 0) }
func (r *sidebarBGRenderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.rect} }
func (r *sidebarBGRenderer) Destroy()                     {}

func (r *sidebarBGRenderer) Refresh() {
	r.rect.FillColor = theme.OverlayBackgroundColor()
	r.rect.Refresh()
}

// buildNav constructs the sidebar + content area layout.
func (a *App) buildNav() fyne.CanvasObject {
	content := container.NewStack()

	dashboard := screens.NewDashboard(a.filterState, a.worklogState, a.repo, a.fyneApp.Preferences(), a.window, a.tr)
	report := screens.NewReport(a.filterState, a.reportState, a.repo, a.fyneApp.Preferences(), a.window)
	manageTeams := screens.NewManageTeams(a.repo, a.tr, a.window)
	settings := screens.NewSettings(a.repo, a.fyneApp.Preferences(), a.tr, a.window, func() {
		a.window.SetContent(a.buildNav())
	})

	worklogPlaceholder := container.NewCenter(widget.NewLabel(a.tr.T("nav.worklog")))
	userOverviewPlaceholder := container.NewCenter(widget.NewLabel(a.tr.T("nav.user_overview")))

	showScreen := func(o fyne.CanvasObject) {
		content.Objects = []fyne.CanvasObject{o}
		content.Refresh()
	}

	var navBtns []*widget.Button

	makeNavBtn := func(label string, screen fyne.CanvasObject) *widget.Button {
		btn := widget.NewButton(label, nil)
		btn.Alignment = widget.ButtonAlignLeading
		btn.Importance = widget.LowImportance
		btn.OnTapped = func() {
			for _, b := range navBtns {
				b.Importance = widget.LowImportance
				b.Refresh()
			}
			btn.Importance = widget.MediumImportance
			btn.Refresh()
			showScreen(screen)
		}
		navBtns = append(navBtns, btn)
		return btn
	}

	dashBtn := makeNavBtn(a.tr.T("nav.dashboard"), dashboard.Canvas())
	worklogBtn := makeNavBtn(a.tr.T("nav.worklog"), worklogPlaceholder)
	userBtn := makeNavBtn(a.tr.T("nav.user_overview"), userOverviewPlaceholder)
	teamsBtn := makeNavBtn(a.tr.T("nav.teams"), manageTeams.Canvas())
	reportsBtn := makeNavBtn(a.tr.T("nav.reports"), report.Canvas())
	jiraConfigBtn := makeNavBtn(a.tr.T("nav.jira_config"), settings.Canvas())

	accordion := widget.NewAccordion(
		widget.NewAccordionItem(
			a.tr.T("nav.section.main"),
			container.NewVBox(dashBtn, worklogBtn, userBtn, teamsBtn, reportsBtn),
		),
		widget.NewAccordionItem(
			a.tr.T("nav.settings"),
			container.NewVBox(jiraConfigBtn),
		),
	)
	accordion.MultiOpen = true
	accordion.Open(0)
	accordion.Open(1)

	dashBtn.Importance = widget.MediumImportance
	showScreen(dashboard.Canvas())

	sidebar := container.NewStack(newThemedSidebarBG(), container.NewPadded(accordion))
	return container.New(&sidebarLayout{width: 240}, sidebar, content)
}
