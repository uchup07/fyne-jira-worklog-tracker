# Plan 4: Polish + Ship — i18n, Export, Teams/Settings, fyne package

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire internationalization (EN/ID) throughout the app, implement CSV/Excel/PDF export with file-save dialogs, complete the Manage Teams and Settings screens, finalize the custom theme, and produce a distributable `.app` bundle with `fyne package`.

**Architecture:** `i18n.I18n` is passed by pointer to every screen and widget that renders user-facing strings. `T(key)` replaces all hardcoded strings. Export functions in `export/` take typed data structs and a file path — screens call `dialog.ShowFileSave()` to let the user choose the destination. Language switching calls `i18n.SetLang()` and triggers a full window refresh.

**Prerequisites:** Plans 1–3 complete — all screens compile and the app runs with real Jira data.

**Tech Stack:** `go-i18n/v2`, `embed`, `encoding/csv`, `excelize/v2`, `gofpdf`, `fyne package`

---

## File Map

| File | Action | Purpose |
|---|---|---|
| `i18n/i18n.go` | Create | T() function, SetLang(), embed directives |
| `i18n/locales/en.json` | Create | English string keys |
| `i18n/locales/id.json` | Create | Indonesian string keys |
| `export/csv.go` | Create | CSV export for WorklogGroup data |
| `export/excel.go` | Create | Excel export using excelize/v2 |
| `export/pdf.go` | Create | PDF export using gofpdf |
| `export/export_test.go` | Create | Tests for all three export formats |
| `screens/manage_teams.go` | Replace | Full team/member CRUD screen |
| `screens/settings.go` | Replace | Full settings: Jira config + holidays + language |
| `custom/theme.go` | Modify | Complete theme: spacing, fonts, icon overrides |
| `app/app.go` | Modify | Pass i18n to all screens |
| `app/nav.go` | Modify | Pass i18n to all screens |
| `FyneApp.toml` | Create | App metadata for fyne package |

---

## Task 1: i18n — T() Function + Locale Files

**Files:**
- Create: `i18n/i18n.go`
- Create: `i18n/locales/en.json`
- Create: `i18n/locales/id.json`

- [ ] **Step 1: Create `i18n/locales/en.json`**

```json
{
  "app.title": "Jira Worklog Tracker",
  "nav.dashboard": "Dashboard",
  "nav.report": "Report",
  "nav.teams": "Teams",
  "nav.settings": "Settings",

  "setup.title": "Welcome to Jira Worklog Tracker",
  "setup.subtitle": "Enter your Jira credentials to get started.",
  "setup.field.domain": "Jira Domain",
  "setup.field.email": "Email",
  "setup.field.token": "API Token",
  "setup.field.workRefField": "Work Ref Field ID",
  "setup.btn.test": "Test Connection",
  "setup.btn.save": "Save & Continue",
  "setup.status.testing": "Testing...",
  "setup.status.success": "✓ Connection successful",
  "setup.status.failed": "Connection failed: {error}",
  "setup.error.required": "Domain, email, and API token are required",

  "dashboard.filter.from": "From:",
  "dashboard.filter.to": "To:",
  "dashboard.btn.search": "Search",
  "dashboard.btn.cancel": "Cancel",
  "dashboard.tab.worklog": "Work Reference",
  "dashboard.tab.diagram": "Diagram",
  "dashboard.tab.timesheet": "Timesheet",
  "dashboard.progress.searching": "Searching... {found} issues found",
  "dashboard.progress.processing": "Processing {done} / {total}",
  "dashboard.progress.finalizing": "Finalizing results...",
  "dashboard.col.workref": "Work Reference",
  "dashboard.col.hours": "Total Hours",
  "dashboard.col.entries": "Entries",
  "dashboard.col.authors": "Authors",
  "dashboard.col.lastEntry": "Last Entry",
  "dashboard.export.csv": "Export CSV",
  "dashboard.export.excel": "Export Excel",
  "dashboard.export.pdf": "Export PDF",
  "dashboard.error.noConfig": "No Jira configuration found — please complete Setup",
  "dashboard.error.searchFailed": "Search failed: {error}",

  "report.tab.table": "Mapping Table",
  "report.tab.charts": "Charts",
  "report.col.irqKey": "IRQ Key",
  "report.col.jswKey": "JSW Key",
  "report.col.taskName": "Task Name",
  "report.col.vertical": "Vertical",
  "report.col.company": "Company",
  "report.col.hours": "Hours",
  "report.chart.vertical.pie": "Vertical (Pie)",
  "report.chart.vertical.bar": "Vertical (Bar)",
  "report.error.failed": "Report failed: {error}",
  "report.placeholder.charts": "Run a search to see charts",
  "report.export.csv": "Export CSV",
  "report.export.excel": "Export Excel",
  "report.export.pdf": "Export PDF",

  "teams.title": "Manage Teams",
  "teams.btn.addTeam": "Add Team",
  "teams.btn.deleteTeam": "Delete Team",
  "teams.btn.addMember": "Add Member",
  "teams.btn.removeMember": "Remove",
  "teams.col.user": "User",
  "teams.col.joinDate": "Join Date",
  "teams.col.leaveDate": "Leave Date",
  "teams.dialog.newTeam": "New Team Name",
  "teams.dialog.addMember": "Add Member",

  "settings.title": "Settings",
  "settings.section.jira": "Jira Configuration",
  "settings.field.domain": "Jira Domain",
  "settings.field.email": "Email",
  "settings.field.token": "API Token",
  "settings.field.workRefField": "Work Ref Field ID",
  "settings.field.verticalField": "Vertical Field ID",
  "settings.field.companyField": "Company Field ID",
  "settings.field.uatEndField": "UAT End Field ID",
  "settings.btn.save": "Save",
  "settings.btn.testConnection": "Test Connection",
  "settings.section.holidays": "Public Holidays",
  "settings.btn.addHoliday": "Add Holiday",
  "settings.btn.deleteHoliday": "Delete",
  "settings.section.language": "Language",
  "settings.lang.en": "English",
  "settings.lang.id": "Indonesian",
  "settings.saved": "Settings saved successfully",

  "export.dialog.save": "Save File",
  "export.success": "File saved successfully",
  "export.error": "Export failed: {error}"
}
```

- [ ] **Step 2: Create `i18n/locales/id.json`**

```json
{
  "app.title": "Jira Worklog Tracker",
  "nav.dashboard": "Dasbor",
  "nav.report": "Laporan",
  "nav.teams": "Tim",
  "nav.settings": "Pengaturan",

  "setup.title": "Selamat Datang di Jira Worklog Tracker",
  "setup.subtitle": "Masukkan kredensial Jira Anda untuk memulai.",
  "setup.field.domain": "Domain Jira",
  "setup.field.email": "Email",
  "setup.field.token": "Token API",
  "setup.field.workRefField": "ID Field Work Ref",
  "setup.btn.test": "Tes Koneksi",
  "setup.btn.save": "Simpan & Lanjutkan",
  "setup.status.testing": "Menguji...",
  "setup.status.success": "✓ Koneksi berhasil",
  "setup.status.failed": "Koneksi gagal: {error}",
  "setup.error.required": "Domain, email, dan token API diperlukan",

  "dashboard.filter.from": "Dari:",
  "dashboard.filter.to": "Sampai:",
  "dashboard.btn.search": "Cari",
  "dashboard.btn.cancel": "Batal",
  "dashboard.tab.worklog": "Referensi Kerja",
  "dashboard.tab.diagram": "Diagram",
  "dashboard.tab.timesheet": "Timesheet",
  "dashboard.progress.searching": "Mencari... {found} isu ditemukan",
  "dashboard.progress.processing": "Memproses {done} / {total}",
  "dashboard.progress.finalizing": "Menyelesaikan hasil...",
  "dashboard.col.workref": "Referensi Kerja",
  "dashboard.col.hours": "Total Jam",
  "dashboard.col.entries": "Entri",
  "dashboard.col.authors": "Penulis",
  "dashboard.col.lastEntry": "Entri Terakhir",
  "dashboard.export.csv": "Ekspor CSV",
  "dashboard.export.excel": "Ekspor Excel",
  "dashboard.export.pdf": "Ekspor PDF",
  "dashboard.error.noConfig": "Konfigurasi Jira tidak ditemukan — selesaikan Setup terlebih dahulu",
  "dashboard.error.searchFailed": "Pencarian gagal: {error}",

  "report.tab.table": "Tabel Pemetaan",
  "report.tab.charts": "Grafik",
  "report.col.irqKey": "Kunci IRQ",
  "report.col.jswKey": "Kunci JSW",
  "report.col.taskName": "Nama Tugas",
  "report.col.vertical": "Vertikal",
  "report.col.company": "Perusahaan",
  "report.col.hours": "Jam",
  "report.chart.vertical.pie": "Vertikal (Pie)",
  "report.chart.vertical.bar": "Vertikal (Bar)",
  "report.error.failed": "Laporan gagal: {error}",
  "report.placeholder.charts": "Jalankan pencarian untuk melihat grafik",
  "report.export.csv": "Ekspor CSV",
  "report.export.excel": "Ekspor Excel",
  "report.export.pdf": "Ekspor PDF",

  "teams.title": "Kelola Tim",
  "teams.btn.addTeam": "Tambah Tim",
  "teams.btn.deleteTeam": "Hapus Tim",
  "teams.btn.addMember": "Tambah Anggota",
  "teams.btn.removeMember": "Hapus",
  "teams.col.user": "Pengguna",
  "teams.col.joinDate": "Tanggal Bergabung",
  "teams.col.leaveDate": "Tanggal Keluar",
  "teams.dialog.newTeam": "Nama Tim Baru",
  "teams.dialog.addMember": "Tambah Anggota",

  "settings.title": "Pengaturan",
  "settings.section.jira": "Konfigurasi Jira",
  "settings.field.domain": "Domain Jira",
  "settings.field.email": "Email",
  "settings.field.token": "Token API",
  "settings.field.workRefField": "ID Field Work Ref",
  "settings.field.verticalField": "ID Field Vertikal",
  "settings.field.companyField": "ID Field Perusahaan",
  "settings.field.uatEndField": "ID Field UAT End",
  "settings.btn.save": "Simpan",
  "settings.btn.testConnection": "Tes Koneksi",
  "settings.section.holidays": "Hari Libur",
  "settings.btn.addHoliday": "Tambah Hari Libur",
  "settings.btn.deleteHoliday": "Hapus",
  "settings.section.language": "Bahasa",
  "settings.lang.en": "Inggris",
  "settings.lang.id": "Indonesia",
  "settings.saved": "Pengaturan berhasil disimpan",

  "export.dialog.save": "Simpan File",
  "export.success": "File berhasil disimpan",
  "export.error": "Ekspor gagal: {error}"
}
```

- [ ] **Step 3: Create `i18n/i18n.go`**

```go
// i18n/i18n.go
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed locales/*.json
var localesFS embed.FS

// I18n holds the active language strings. Pass *I18n to every screen and widget.
type I18n struct {
	lang    string
	strings map[string]string
}

// New loads the locale file for lang ("en" or "id").
// Falls back to English if the locale file is missing.
func New(lang string) *I18n {
	i := &I18n{}
	i.load(lang)
	return i
}

func (i *I18n) load(lang string) {
	data, err := localesFS.ReadFile(fmt.Sprintf("locales/%s.json", lang))
	if err != nil {
		// Fallback to English
		data, _ = localesFS.ReadFile("locales/en.json")
		lang = "en"
	}
	var m map[string]string
	json.Unmarshal(data, &m)
	i.lang = lang
	i.strings = m
}

// Lang returns the currently active language code.
func (i *I18n) Lang() string { return i.lang }

// T looks up key and substitutes any {param} placeholders from params.
// If the key is not found, the key itself is returned (useful for debugging).
//
// Usage:
//
//	i18n.T("dashboard.progress.searching", map[string]any{"found": 42})
//	// → "Searching... 42 issues found"
func (i *I18n) T(key string, params ...map[string]any) string {
	s, ok := i.strings[key]
	if !ok {
		return key // key not found — return raw key so missing strings are obvious
	}
	if len(params) > 0 {
		for k, v := range params[0] {
			s = strings.ReplaceAll(s, fmt.Sprintf("{%s}", k), fmt.Sprintf("%v", v))
		}
	}
	return s
}

// SetLang reloads strings for the given language code.
// Call window.Content().Refresh() after this to repaint all labels.
func (i *I18n) SetLang(lang string) {
	i.load(lang)
}
```

- [ ] **Step 4: Write i18n tests**

```go
// i18n/i18n_test.go
package i18n_test

import (
	"testing"

	"github.com/uchup07/fyne-jira-worklog-tracker/i18n"
)

func TestTReturnsEnglishString(t *testing.T) {
	tr := i18n.New("en")
	got := tr.T("nav.dashboard")
	if got != "Dashboard" {
		t.Errorf("got %q, want %q", got, "Dashboard")
	}
}

func TestTReturnsIndonesianString(t *testing.T) {
	tr := i18n.New("id")
	got := tr.T("nav.dashboard")
	if got != "Dasbor" {
		t.Errorf("got %q, want %q", got, "Dasbor")
	}
}

func TestTSubstitutesParams(t *testing.T) {
	tr := i18n.New("en")
	got := tr.T("dashboard.progress.searching", map[string]any{"found": 42})
	want := "Searching... 42 issues found"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTReturnsMissingKeyAsKey(t *testing.T) {
	tr := i18n.New("en")
	got := tr.T("nonexistent.key")
	if got != "nonexistent.key" {
		t.Errorf("expected key passthrough, got %q", got)
	}
}

func TestSetLangSwitches(t *testing.T) {
	tr := i18n.New("en")
	tr.SetLang("id")
	got := tr.T("nav.dashboard")
	if got != "Dasbor" {
		t.Errorf("after SetLang(id): got %q, want %q", got, "Dasbor")
	}
}
```

- [ ] **Step 5: Run i18n tests**

```bash
go test ./i18n/... -v
```
Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add i18n/
git commit -m "feat: i18n with English/Indonesian locale files + embed + tests"
```

---

## Task 2: Wire i18n into App

**Files:**
- Modify: `app/app.go` (add i18n field)
- Modify: `app/nav.go` (pass i18n to screens)

The i18n pointer is added to `App` and propagated to every screen constructor. Screens that already have string literals will be updated in subsequent tasks as each screen is touched.

- [ ] **Step 1: Add i18n to `app/app.go`**

```go
// In app/app.go — add the i18n field and load it in New():

import (
    ...
    "github.com/uchup07/fyne-jira-worklog-tracker/i18n"
    ...
)

type App struct {
    fyneApp fyne.App
    window  fyne.Window

    filterState  *state.FilterState
    worklogState *state.WorklogState
    reportState  *state.ReportState

    repo *db.Repository
    tr   *i18n.I18n
}

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
```

- [ ] **Step 2: Build to verify**

```bash
go build ./...
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add app/app.go
git commit -m "feat: add i18n to App struct, load language from Preferences"
```

---

## Task 3: Export Functions + Tests

**Files:**
- Create: `export/csv.go`
- Create: `export/excel.go`
- Create: `export/pdf.go`
- Create: `export/export_test.go`

- [ ] **Step 1: Write failing tests first**

```go
// export/export_test.go
package export_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/uchup07/fyne-jira-worklog-tracker/export"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
)

var sampleGroups = []jira.WorklogGroup{
	{
		WorkReference: "CR-001",
		TotalSeconds:  10800,
		Items: []jira.WorklogItem{
			{IssueKey: "JSW-1", IssueSummary: "Task A", Author: jira.User{DisplayName: "Alice"}, TimeSpentSeconds: 7200, Started: time.Date(2026, 1, 15, 9, 0, 0, 0, time.UTC)},
			{IssueKey: "JSW-2", IssueSummary: "Task B", Author: jira.User{DisplayName: "Bob"}, TimeSpentSeconds: 3600, Started: time.Date(2026, 1, 15, 14, 0, 0, 0, time.UTC)},
		},
	},
}

func TestWriteCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	if err := export.WriteCSV(path, sampleGroups); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}
	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Error("expected non-empty CSV file")
	}
	content := string(data)
	if !contains(content, "CR-001") {
		t.Error("CSV missing work reference")
	}
	if !contains(content, "Alice") {
		t.Error("CSV missing author name")
	}
}

func TestWriteExcel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.xlsx")
	if err := export.WriteExcel(path, sampleGroups); err != nil {
		t.Fatalf("WriteExcel: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() == 0 {
		t.Error("expected non-empty xlsx file")
	}
}

func TestWritePDF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.pdf")
	if err := export.WritePDF(path, sampleGroups, "2026-01-01", "2026-01-31"); err != nil {
		t.Fatalf("WritePDF: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() == 0 {
		t.Error("expected non-empty PDF file")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./export/... -v
```
Expected: compilation failure.

- [ ] **Step 3: Create `export/csv.go`**

```go
// export/csv.go
package export

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
)

// WriteCSV writes worklog groups to a CSV file at path.
func WriteCSV(path string, groups []jira.WorklogGroup) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create CSV: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Header
	w.Write([]string{"Work Reference", "Issue Key", "Summary", "Author", "Hours", "Date", "Comment"})

	for _, group := range groups {
		for _, item := range group.Items {
			w.Write([]string{
				group.WorkReference,
				item.IssueKey,
				item.IssueSummary,
				item.Author.DisplayName,
				fmt.Sprintf("%.2f", float64(item.TimeSpentSeconds)/3600),
				item.Started.Format("2006-01-02"),
				item.Comment,
			})
		}
	}
	return w.Error()
}
```

- [ ] **Step 4: Create `export/excel.go`**

```go
// export/excel.go
package export

import (
	"fmt"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/xuri/excelize/v2"
)

// WriteExcel writes worklog groups to an Excel (.xlsx) file at path.
func WriteExcel(path string, groups []jira.WorklogGroup) error {
	f := excelize.NewFile()
	sheet := "Worklogs"
	f.SetSheetName("Sheet1", sheet)

	// Header row with bold style
	bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	headers := []string{"Work Reference", "Issue Key", "Summary", "Author", "Hours", "Date", "Comment"}
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, bold)
	}

	row := 2
	for _, group := range groups {
		for _, item := range group.Items {
			values := []any{
				group.WorkReference,
				item.IssueKey,
				item.IssueSummary,
				item.Author.DisplayName,
				fmt.Sprintf("%.2f", float64(item.TimeSpentSeconds)/3600),
				item.Started.Format("2006-01-02"),
				item.Comment,
			}
			for col, v := range values {
				cell, _ := excelize.CoordinatesToCellName(col+1, row)
				f.SetCellValue(sheet, cell, v)
			}
			row++
		}
	}

	// Auto-fit columns A–G
	for col := 1; col <= 7; col++ {
		name, _ := excelize.ColumnNumberToName(col)
		f.SetColWidth(sheet, name, name, 18)
	}

	return f.SaveAs(path)
}
```

- [ ] **Step 5: Create `export/pdf.go`**

```go
// export/pdf.go
package export

import (
	"fmt"

	"github.com/jung-kurt/gofpdf"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
)

// WritePDF writes a worklog summary report to a PDF file at path.
func WritePDF(path string, groups []jira.WorklogGroup, startDate, endDate string) error {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Jira Worklog Report")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 8, fmt.Sprintf("Period: %s to %s", startDate, endDate))
	pdf.Ln(12)

	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(37, 99, 235)
	pdf.SetTextColor(255, 255, 255)
	colWidths := []float64{50, 30, 60, 40, 25, 30, 55}
	headers := []string{"Work Ref", "Issue", "Summary", "Author", "Hours", "Date", "Comment"}
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(0, 0, 0)
	fill := false
	for _, group := range groups {
		for _, item := range group.Items {
			if fill {
				pdf.SetFillColor(235, 241, 255)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}
			values := []string{
				group.WorkReference,
				item.IssueKey,
				truncate(item.IssueSummary, 28),
				item.Author.DisplayName,
				fmt.Sprintf("%.1f", float64(item.TimeSpentSeconds)/3600),
				item.Started.Format("2006-01-02"),
				truncate(item.Comment, 28),
			}
			for i, v := range values {
				pdf.CellFormat(colWidths[i], 7, v, "1", 0, "L", fill, 0, "")
			}
			pdf.Ln(-1)
			fill = !fill
		}
	}

	// Summary footer
	pdf.Ln(4)
	pdf.SetFont("Arial", "B", 10)
	totalHours := 0.0
	for _, g := range groups {
		totalHours += float64(g.TotalSeconds) / 3600
	}
	pdf.Cell(0, 8, fmt.Sprintf("Total: %.1f hours across %d work references", totalHours, len(groups)))

	return pdf.OutputFileAndClose(path)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
```

- [ ] **Step 6: Run the export tests**

```bash
go test ./export/... -v
```
Expected: all three tests pass.

- [ ] **Step 7: Commit**

```bash
git add export/ i18n/
git commit -m "feat: CSV/Excel/PDF export functions + i18n T() function"
```

---

## Task 4: Manage Teams Screen (Full Implementation)

**Files:**
- Replace: `screens/manage_teams.go`

- [ ] **Step 1: Replace `screens/manage_teams.go`**

```go
// screens/manage_teams.go
package screens

import (
	"fmt"

	"github.com/uchup07/fyne-jira-worklog-tracker/db"
	"github.com/uchup07/fyne-jira-worklog-tracker/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ManageTeams is the team and team-member CRUD screen.
type ManageTeams struct {
	repo      *db.Repository
	tr        *i18n.I18n
	window    fyne.Window
	canvas    fyne.CanvasObject

	teams      []db.Team
	members    []db.TeamMember
	selectedID int // currently selected team ID

	teamList    *widget.List
	memberTable *widget.Table
}

// NewManageTeams creates the Manage Teams screen.
func NewManageTeams(repo *db.Repository, tr *i18n.I18n, window fyne.Window) *ManageTeams {
	m := &ManageTeams{repo: repo, tr: tr, window: window}
	m.build()
	m.reload()
	return m
}

// Canvas returns the Fyne canvas object.
func (m *ManageTeams) Canvas() fyne.CanvasObject { return m.canvas }

func (m *ManageTeams) build() {
	// Left panel: team list + Add/Delete
	m.teamList = widget.NewList(
		func() int { return len(m.teams) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(m.teams) {
				item.(*widget.Label).SetText(m.teams[id].Name)
			}
		},
	)
	m.teamList.OnSelected = func(id widget.ListItemID) {
		if id < len(m.teams) {
			m.selectedID = m.teams[id].ID
			m.reloadMembers()
		}
	}

	addTeamBtn := widget.NewButton(m.tr.T("teams.btn.addTeam"), m.showAddTeamDialog)
	delTeamBtn := widget.NewButton(m.tr.T("teams.btn.deleteTeam"), m.deleteSelectedTeam)

	leftPanel := container.NewBorder(
		nil,
		container.NewHBox(addTeamBtn, delTeamBtn),
		nil, nil,
		m.teamList,
	)

	// Right panel: member table + Add/Remove
	cols := []string{m.tr.T("teams.col.user"), m.tr.T("teams.col.joinDate"), m.tr.T("teams.col.leaveDate"), ""}
	colWidths := []float32{200, 100, 100, 80}

	m.memberTable = widget.NewTable(
		func() (int, int) { return len(m.members), len(cols) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel(""))
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			if id.Row >= len(m.members) {
				return
			}
			mem := m.members[id.Row]
			box := cell.(*fyne.Container)
			switch id.Col {
			case 0:
				setBoxLabel(box, mem.UserID)
			case 1:
				setBoxLabel(box, mem.JoinDate)
			case 2:
				setBoxLabel(box, mem.LeaveDate)
			case 3:
				btn := widget.NewButton(m.tr.T("teams.btn.removeMember"), func() {
					m.repo.Teams.RemoveMember(mem.ID)
					m.reloadMembers()
				})
				box.Objects = []fyne.CanvasObject{btn}
				box.Refresh()
			}
		},
	)
	for i, w := range colWidths {
		m.memberTable.SetColumnWidth(i, w)
	}

	addMemberBtn := widget.NewButton(m.tr.T("teams.btn.addMember"), m.showAddMemberDialog)
	rightPanel := container.NewBorder(
		nil,
		addMemberBtn,
		nil, nil,
		m.memberTable,
	)

	m.canvas = container.NewHSplit(leftPanel, rightPanel)
}

func (m *ManageTeams) reload() {
	teams, err := m.repo.Teams.ListTeams()
	if err != nil {
		return
	}
	m.teams = teams
	m.teamList.Refresh()
}

func (m *ManageTeams) reloadMembers() {
	if m.selectedID == 0 {
		m.members = nil
		m.memberTable.Refresh()
		return
	}
	members, err := m.repo.Teams.ListMembers(m.selectedID)
	if err != nil {
		return
	}
	m.members = members
	m.memberTable.Refresh()
}

func (m *ManageTeams) showAddTeamDialog() {
	entry := widget.NewEntry()
	entry.PlaceHolder = m.tr.T("teams.dialog.newTeam")
	dialog.ShowForm(m.tr.T("teams.btn.addTeam"), "Add", "Cancel",
		[]*widget.FormItem{widget.NewFormItem("Name", entry)},
		func(ok bool) {
			if !ok || entry.Text == "" {
				return
			}
			if _, err := m.repo.Teams.CreateTeam(entry.Text); err != nil {
				dialog.ShowError(fmt.Errorf("create team: %w", err), m.window)
				return
			}
			m.reload()
		}, m.window)
}

func (m *ManageTeams) deleteSelectedTeam() {
	if m.selectedID == 0 {
		return
	}
	dialog.ShowConfirm("Delete Team", "Delete this team and all its members?", func(ok bool) {
		if !ok {
			return
		}
		if err := m.repo.Teams.DeleteTeam(m.selectedID); err != nil {
			dialog.ShowError(fmt.Errorf("delete team: %w", err), m.window)
			return
		}
		m.selectedID = 0
		m.members = nil
		m.memberTable.Refresh()
		m.reload()
	}, m.window)
}

func (m *ManageTeams) showAddMemberDialog() {
	if m.selectedID == 0 {
		dialog.ShowInformation("No Team Selected", "Please select a team first.", m.window)
		return
	}
	userEntry := widget.NewEntry()
	userEntry.PlaceHolder = "Jira accountId"
	joinEntry := widget.NewEntry()
	joinEntry.PlaceHolder = "2006-01-02"
	leaveEntry := widget.NewEntry()
	leaveEntry.PlaceHolder = "2006-01-02 (optional)"

	dialog.ShowForm(m.tr.T("teams.btn.addMember"), "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem(m.tr.T("teams.col.user"), userEntry),
			widget.NewFormItem(m.tr.T("teams.col.joinDate"), joinEntry),
			widget.NewFormItem(m.tr.T("teams.col.leaveDate"), leaveEntry),
		},
		func(ok bool) {
			if !ok || userEntry.Text == "" {
				return
			}
			err := m.repo.Teams.AddMember(db.TeamMember{
				TeamID:    m.selectedID,
				UserID:    userEntry.Text,
				JoinDate:  joinEntry.Text,
				LeaveDate: leaveEntry.Text,
			})
			if err != nil {
				dialog.ShowError(fmt.Errorf("add member: %w", err), m.window)
				return
			}
			m.reloadMembers()
		}, m.window)
}

func setBoxLabel(box *fyne.Container, text string) {
	if len(box.Objects) == 0 {
		box.Objects = []fyne.CanvasObject{widget.NewLabel(text)}
	} else {
		box.Objects[0].(*widget.Label).SetText(text)
	}
	box.Refresh()
}
```

- [ ] **Step 2: Update `app/nav.go` to pass deps to ManageTeams**

```go
manageTeams := screens.NewManageTeams(a.repo, a.tr, a.window)
```

- [ ] **Step 3: Build**

```bash
go build ./...
```

---

## Task 5: Settings Screen (Full Implementation)

**Files:**
- Replace: `screens/settings.go`

- [ ] **Step 1: Replace `screens/settings.go`**

```go
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

// Settings shows Jira config, holidays, and language switcher.
type Settings struct {
	repo      *db.Repository
	prefs     fyne.Preferences
	tr        *i18n.I18n
	window    fyne.Window
	onLangChange func() // called after language switch so window repaints
	canvas    fyne.CanvasObject
}

// NewSettings creates the Settings screen.
// onLangChange is called when the user switches language — pass window.Content().Refresh.
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
	options := []string{s.tr.T("settings.lang.en"), s.tr.T("settings.lang.id")}
	selected := options[0]
	if current == "id" {
		selected = options[1]
	}

	radio := widget.NewRadioGroup(options, func(choice string) {
		lang := "en"
		if choice == options[1] {
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
```

- [ ] **Step 2: Update `app/nav.go` to pass deps to Settings**

In `app/nav.go`:
```go
settings := screens.NewSettings(a.repo, a.fyneApp.Preferences(), a.tr, a.window, func() {
    a.window.Content().Refresh()
})
```

- [ ] **Step 3: Build**

```bash
go build ./...
```

---

## Task 6: Add Export Buttons to Dashboard and Report

**Files:**
- Modify: `screens/dashboard.go`
- Modify: `screens/report.go`

- [ ] **Step 1: Add export buttons to `screens/dashboard.go`**

In the `build()` method of Dashboard, after building `tabs`, add an export toolbar:

```go
// Add below the tabs definition in dashboard.build():

exportBar := container.NewHBox(
    widget.NewButton(d.tr.T("dashboard.export.csv"), func() {
        dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
            if err != nil || uri == nil {
                return
            }
            path := uri.URI().Path()
            uri.Close()
            groups := collectGroups(d.ws)
            if err := export.WriteCSV(path, groups); err != nil {
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
            groups := collectGroups(d.ws)
            if err := export.WriteExcel(path, groups); err != nil {
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
            groups := collectGroups(d.ws)
            start, _ := d.ws.SearchStart.Get()
            end, _ := d.ws.SearchEnd.Get()
            if err := export.WritePDF(path, groups, start, end); err != nil {
                dialog.ShowError(fmt.Errorf("PDF export: %w", err), d.window)
            }
        }, d.window)
    }),
)
```

Add `collectGroups` helper at the bottom of `screens/dashboard.go`:

```go
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
```

Also add the `export` import to `screens/dashboard.go`:
```go
"github.com/uchup07/fyne-jira-worklog-tracker/export"
```

Update `d.canvas` to include the export bar:
```go
top := container.NewVBox(filterBar.Canvas(), d.progress.Canvas(), exportBar)
d.canvas = container.NewBorder(top, nil, nil, nil, tabs)
```

You'll also need to add `tr *i18n.I18n` to Dashboard and pass it through. Update `NewDashboard` signature:
```go
func NewDashboard(fs *state.FilterState, ws *state.WorklogState, repo *db.Repository, prefs fyne.Preferences, window fyne.Window, tr *i18n.I18n) *Dashboard {
```

And update `app/nav.go`:
```go
dashboard := screens.NewDashboard(a.filterState, a.worklogState, a.repo, a.fyneApp.Preferences(), a.window, a.tr)
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

---

## Task 7: FyneApp.toml + Build + Final Checkpoint

**Files:**
- Create: `FyneApp.toml`

- [ ] **Step 1: Create `FyneApp.toml`**

```toml
Website = "https://github.com/uchup07/fyne-jira-worklog-tracker"
Email   = "uchup07@gmail.com"

[Details]
  Icon      = "assets/icon.png"
  Name      = "Jira Worklog Tracker"
  ID        = "com.uchup07.jira-worklog-tracker"
  Version   = "1.0.0"
  Build     = 1

  [Details.darwin]
    Category = "public.app-category.productivity"
```

- [ ] **Step 2: Install fyne CLI tool**

```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
```
Expected: `fyne` binary appears in `$(go env GOPATH)/bin`.

- [ ] **Step 3: Full build + all tests**

```bash
go build ./...
go test ./...
```
Expected: all tests pass, clean build.

- [ ] **Step 4: Run the full app — visual acceptance test**

```bash
go run .
```

Verify all success criteria from the design spec:
- [ ] App launches without a server; sidebar nav works
- [ ] First run shows Setup wizard; config saves and transitions to Dashboard
- [ ] Dashboard search with a real date range fetches and displays worklogs
- [ ] Progress bar animates; Cancel stops the in-flight request
- [ ] Report screen renders mapping table and charts
- [ ] Teams screen: create a team, add a member, remove a member, delete team
- [ ] Settings screen: edit config, switch language to Indonesian, verify labels change throughout app, switch back
- [ ] Dashboard export buttons open a file-save dialog; saved files are non-empty
- [ ] All three export formats (CSV, Excel, PDF) produce valid files

- [ ] **Step 5: Package the macOS app bundle**

```bash
fyne package -os darwin -icon assets/icon.png -name "Jira Worklog Tracker"
```
Expected: `Jira Worklog Tracker.app` created in the current directory. Open it from Finder — it should run without a terminal.

- [ ] **Step 6: Final commit and push**

```bash
git add .
git commit -m "$(cat <<'EOF'
feat: complete app — i18n EN/ID, CSV/Excel/PDF export, Teams/Settings screens, fyne package

- i18n with embed + T() substitution, SetLang() + window refresh
- export/csv, export/excel, export/pdf with file-save dialogs
- ManageTeams: team CRUD + member add/remove
- Settings: Jira config edit, holiday manager, language switcher
- FyneApp.toml for fyne package build

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
git push origin main
```

---

## Phase 4 Complete ✓ — App Ships

**All success criteria from the design spec are now met:**
- [ ] Single binary, no server required
- [ ] Setup wizard → Dashboard with live Jira search
- [ ] Report screen with mapping table and charts
- [ ] Teams and holidays CRUD
- [ ] English/Indonesian language switching at runtime
- [ ] CSV, Excel, PDF export with native file-save dialogs
- [ ] `fyne package` produces a distributable `.app` on macOS

**To cross-compile for Windows/Linux:**
```bash
# Requires Docker
go install github.com/fyne-io/fyne-cross/cmd/fyne-cross@latest
fyne-cross windows -arch=amd64
fyne-cross linux   -arch=amd64
```
