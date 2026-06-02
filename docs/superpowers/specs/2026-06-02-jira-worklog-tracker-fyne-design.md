# Jira Worklog Tracker — Fyne.io (Go) Design Spec

**Date:** 2026-06-02  
**Reference:** `~/Public/www/jira-worklog-tracker/FYNE_ARCHITECTURE.md`  
**Status:** Approved — ready for implementation planning

---

## 1. What We're Building

A single Go binary desktop app using Fyne.io v2. No server, no browser — the binary is both the UI and the backend. Jira API calls go directly from the app via `net/http`. Local data lives in SQLite. Credentials are stored in the OS keychain via `fyne.Preferences`.

This is a full port of the existing Next.js Jira Worklog Tracker web app into a native desktop application.

### Goals

- Cross-platform native desktop app (macOS primary, Windows/Linux via `fyne-cross`)
- Single distributable binary — no Docker, no Node.js runtime
- API token stored securely in OS keychain (automatic upgrade from plaintext SQLite)
- Full parity with the web app's core features (worklogs, reports, teams, settings)
- Learning-paced build: Go beginner + Fyne beginner, incremental 8-phase plan

### Out of Scope

- PPTX and Word export (no mature Go library; too complex to port)
- Third-party chart library (`fynesimplechart` is under-maintained — custom canvas instead)
- Server deployment / Docker

---

## 2. Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                  Single Go Binary                    │
│                                                     │
│  ┌──────────┐   ┌──────────┐   ┌────────────────┐  │
│  │  Screens │──▶│  State   │──▶│    Widgets     │  │
│  │ (5 total)│   │ (binding)│   │ (auto-refresh) │  │
│  └────┬─────┘   └──────────┘   └────────────────┘  │
│       │                                             │
│  ┌────▼─────┐   ┌──────────┐   ┌────────────────┐  │
│  │  Jira    │   │    DB    │   │     i18n       │  │
│  │  Client  │   │ SQLite   │   │  go-i18n EN/ID │  │
│  │ net/http │   │ (no CGO) │   │                │  │
│  └────┬─────┘   └──────────┘   └────────────────┘  │
└───────┼─────────────────────────────────────────────┘
        │
   Jira Cloud API
```

**Central pattern:** goroutine runs background work → sends `ProgressEvent` to a channel → reader goroutine calls `fyne.Do()` to update bound widgets on the UI thread. This replaces the entire SSE/polling system from the Next.js app.

---

## 3. Five Screens

| Screen | Trigger | Key widgets |
|---|---|---|
| **Setup** | First run (no config in DB) | `widget.NewForm()`, `widget.NewPasswordEntry()`, test-connection button |
| **Dashboard** | Default after setup | FilterBar, AppTabs (worklogs / diagram / timesheet), ProgressBar |
| **Report** | Sidebar nav | MappingTable, AppTabs, pie + bar charts |
| **Manage Teams** | Sidebar nav | Team list, member table, add/remove buttons, date entries |
| **Settings** | Sidebar nav | Jira config form, holiday list, language radio group |

Navigation is a left sidebar (`container.NewBorder`) switching a central `container.NewStack`.

---

## 4. Project Structure

```
jira-worklog-tracker/
├── main.go
├── go.mod / go.sum
│
├── app/
│   ├── app.go          # App struct — window, state, services, prefs
│   └── nav.go          # Sidebar navigation builder
│
├── screens/
│   ├── setup.go
│   ├── dashboard.go
│   ├── report.go
│   ├── manage_teams.go
│   └── settings.go
│
├── widgets/
│   ├── filter_bar.go
│   ├── worklog_table.go
│   ├── mapping_table.go
│   ├── timesheet.go
│   ├── progress_bar.go
│   ├── pie_chart.go
│   └── bar_chart.go
│
├── state/
│   ├── filter.go       # FilterState (binding.String / UntypedList)
│   ├── worklog.go      # WorklogState
│   └── report.go       # ReportState
│
├── jira/
│   ├── client.go       # net/http + Basic Auth
│   ├── worklogs.go     # FetchWorklogs() + progress channel
│   ├── mapping.go      # BuildMappingReport() + aggregators
│   ├── users.go
│   ├── projects.go
│   └── fields.go
│
├── db/
│   ├── db.go           # SQLite open + inline migration
│   ├── config_repo.go  # API token via fyne.Preferences
│   ├── team_repo.go
│   ├── holiday_repo.go
│   └── user_repo.go
│
├── export/
│   ├── csv.go          # encoding/csv (stdlib)
│   ├── excel.go        # excelize/v2
│   └── pdf.go          # gofpdf
│
├── i18n/
│   ├── i18n.go         # T() + SetLang()
│   ├── locales/en.json
│   └── locales/id.json
│
├── custom/
│   └── theme.go        # AppTheme (blue-600 primary)
│
└── assets/
    ├── icon.png
    └── logo.png
```

---

## 5. Key Technical Decisions

### State Management

Three state structs passed by pointer from `app.App` to each screen:

- `FilterState` — start/end date, selected authors, projects, teams, date mode
- `WorklogState` — worklog groups binding, active tab, search dates
- `ReportState` — mapping report, pre-CR report bindings

All fields are `binding.*` types. Widgets bind directly — `.Set()` from a goroutine triggers automatic widget refresh. No global variables.

### UI Thread Safety

All goroutines that modify UI state must wrap updates in `fyne.Do(func(){...})`. This is structurally enforced: the progress-reader goroutine is the sole caller of `fyne.Do` for any given search, and it's always co-located with the worker goroutine launch in the same screen method.

### Search Cancellation

Each screen that launches a background search holds a `context.CancelFunc`. Clicking Search cancels any prior in-flight context before starting a new one. All `net/http` requests receive the context, so cancellation propagates to the TCP layer automatically.

### Database

- `modernc.org/sqlite` — pure Go SQLite, no CGO required
- Schema managed via inline `CREATE TABLE IF NOT EXISTS` in `db.go` — no migration framework
- API token is **never written to SQLite** — stored in `fyne.Preferences` (OS keychain on macOS)
- All other config fields (domain, email, field IDs) stored in `app_config` table

### i18n

- `go-i18n/v2` with `en.json` and `id.json` locale files embedded via `//go:embed`
- `i18n.I18n` passed by pointer to every screen and widget that renders strings
- `T(key, params...)` signature mirrors the existing Next.js `t()` hook — string keys transfer directly
- Language switch calls `i18n.SetLang()` + `window.Content().Refresh()`

### Charts

Custom `canvas.Raster` widgets using Go's `image/draw` stdlib — no third-party chart dependency. `widgets/pie_chart.go` and `widgets/bar_chart.go` accept `[]float64` data slices and render via a `func(w, h int) image.Image` callback.

### Export

| Format | Library | Notes |
|---|---|---|
| CSV | `encoding/csv` (stdlib) | No dependency |
| Excel | `github.com/xuri/excelize/v2` | Pure Go |
| PDF | `github.com/jung-kurt/gofpdf` | Pure Go |

Export functions live in `export/` and receive typed data structs — no screen-level coupling.

---

## 6. The 8-Phase Build Plan

| Phase | Focus | Checkpoint |
|---|---|---|
| **1** | Go fundamentals (structs, interfaces, goroutines, channels, context) | `go run` exercises pass |
| **2** | Fyne fundamentals: `main.go`, `app.go`, nav skeleton, Setup screen | App window opens, nav switches between placeholder screens |
| **3** | State + data binding: `state/` package | Label auto-updates when binding changes |
| **4** | HTTP + goroutines: `jira/client.go` + `jira/worklogs.go` | Real Jira search runs in background, progress bar updates live, cancel works |
| **5** | Database: `db/` package, all repos | SQLite opens, schema migrates, config saves/loads, token in Preferences |
| **6** | Jira integration: Dashboard screen fully wired | Filter → real search → worklog table populated with live data |
| **7** | Report pipeline: `jira/mapping.go` + Report screen | Mapping table and charts render real data |
| **8** | Polish + export: theme, i18n wired throughout, CSV/Excel/PDF, `fyne package` | Distributable binary produced |

---

## 7. Dependencies (go.mod)

```go
require (
    fyne.io/fyne/v2 v2.6.0
    modernc.org/sqlite v1.34.0
    github.com/xuri/excelize/v2 v2.9.0
    github.com/jung-kurt/gofpdf v1.16.2
    github.com/nicksnyder/go-i18n/v2 v2.4.1
    golang.org/x/text v0.21.0
)
```

CGO is required only for Fyne (OpenGL rendering). The SQLite driver and all export libraries are pure Go. Cross-compilation uses `fyne-cross` Docker toolchain.

---

## 8. Success Criteria

- [ ] App launches without a server; single binary distribution works on macOS
- [ ] Setup screen saves Jira config; subsequent launches go straight to Dashboard
- [ ] Dashboard search fetches real Jira worklogs with live progress updates and working cancel
- [ ] Report screen renders mapping table and charts from real data
- [ ] Teams and holidays can be created, edited, and deleted
- [ ] Language can be switched between English and Indonesian at runtime
- [ ] CSV, Excel, and PDF export produce valid files
- [ ] `fyne package` produces a `.app` bundle on macOS
