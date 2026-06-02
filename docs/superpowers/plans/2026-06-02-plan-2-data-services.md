# Plan 2: Data Services — Jira HTTP Client + SQLite Database Layer

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the Jira API client (HTTP + auth + pagination), the worklog fetch pipeline with goroutine-based progress, and the SQLite database repository layer. By the end, a background Jira search runs, progress fires, cancellation works, and config/teams/holidays persist in SQLite.

**Architecture:** `jira/` package wraps `net/http` with Basic Auth. `jira/worklogs.go` runs on a goroutine, sending `ProgressEvent` structs to a channel. `db/` wraps `database/sql` + `modernc.org/sqlite`; the API token is never written to SQLite — it goes to `fyne.Preferences`. `app.App` is updated to hold the DB and a lazy Jira client.

**Prerequisites:** Plan 1 complete (`go.mod` exists, `app/app.go` compiles).

**Tech Stack:** `net/http`, `net/http/httptest` (tests), `database/sql`, `modernc.org/sqlite`

---

## File Map

| File | Action | Purpose |
|---|---|---|
| `jira/types.go` | Create | All shared Jira domain types |
| `jira/client.go` | Create | HTTP wrapper with Basic Auth |
| `jira/client_test.go` | Create | Mock-server tests for Get/Post/error |
| `jira/worklogs.go` | Create | FetchWorklogs + JQL builder + grouping |
| `jira/worklogs_test.go` | Create | Tests using httptest mock server |
| `jira/users.go` | Create | FetchUsers() |
| `jira/projects.go` | Create | FetchProjects() |
| `jira/fields.go` | Create | FetchFields() |
| `db/db.go` | Create | SQLite open + inline schema migration |
| `db/types.go` | Create | DB-layer domain types |
| `db/config_repo.go` | Create | AppConfig CRUD |
| `db/team_repo.go` | Create | Team + TeamMember CRUD |
| `db/holiday_repo.go` | Create | PublicHoliday CRUD |
| `db/user_repo.go` | Create | JiraUser cache CRUD |
| `db/db_test.go` | Create | In-memory SQLite tests for all repos |
| `app/app.go` | Modify | Add DB repo + Jira client builder |

---

## Task 1: Jira Types

**Files:**
- Create: `jira/types.go`

- [ ] **Step 1: Create `jira/types.go`**

All types the `jira` package shares between files live here.

```go
// jira/types.go
package jira

import "time"

// User represents a Jira account.
type User struct {
	AccountID    string            `json:"accountId"`
	DisplayName  string            `json:"displayName"`
	EmailAddress string            `json:"emailAddress"`
	AvatarUrls   map[string]string `json:"avatarUrls"`
	Active       bool              `json:"active"`
}

// Worklog represents a single worklog entry on a Jira issue.
type Worklog struct {
	ID               string `json:"id"`
	Author           User   `json:"author"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
	Started          string `json:"started"` // "2006-01-02T15:04:05.000+0700"
	Comment          string // flattened from Atlassian Document Format (ADF)
}

// Issue is decoded from the Jira search API. Fields is a raw map so we can
// read both typed fields (summary, project, worklog) and custom fields
// (e.g. "customfield_10001") without a custom unmarshaler.
type Issue struct {
	Key    string         `json:"key"`
	Fields map[string]any `json:"fields"`
}

// Summary returns the issue summary string.
func (i Issue) Summary() string {
	s, _ := i.Fields["summary"].(string)
	return s
}

// ProjectKey returns the Jira project key (e.g. "JSW").
func (i Issue) ProjectKey() string {
	p, _ := i.Fields["project"].(map[string]any)
	k, _ := p["key"].(string)
	return k
}

// ProjectName returns the Jira project display name.
func (i Issue) ProjectName() string {
	p, _ := i.Fields["project"].(map[string]any)
	n, _ := p["name"].(string)
	return n
}

// WorklogTotal returns the total number of worklogs on this issue per the API.
func (i Issue) WorklogTotal() int {
	wl, _ := i.Fields["worklog"].(map[string]any)
	total, _ := wl["total"].(float64)
	return int(total)
}

// InlineWorklogs returns the worklogs already embedded in the search response
// (Jira returns up to 20 inline; more requires a separate paginated call).
func (i Issue) InlineWorklogs() []Worklog {
	wlField, _ := i.Fields["worklog"].(map[string]any)
	wlsRaw, _ := wlField["worklogs"].([]any)
	return decodeWorklogs(wlsRaw)
}

// CustomField reads a string value from a custom field by field ID.
// Handles both plain strings and objects with a "value" or "name" key.
func (i Issue) CustomField(fieldID string) string {
	v, ok := i.Fields[fieldID]
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case map[string]any:
		if s, ok := val["value"].(string); ok {
			return s
		}
		if s, ok := val["name"].(string); ok {
			return s
		}
	}
	return ""
}

// SearchResponse is the response from POST /rest/api/3/search/jql.
type SearchResponse struct {
	Issues        []Issue `json:"issues"`
	NextPageToken string  `json:"nextPageToken"`
	Total         int     `json:"total"`
}

// WorklogListResponse is the response from GET /rest/api/3/issue/{key}/worklog.
type WorklogListResponse struct {
	Worklogs   []Worklog `json:"worklogs"`
	Total      int       `json:"total"`
	MaxResults int       `json:"maxResults"`
	StartAt    int       `json:"startAt"`
}

// SearchFilters holds all filter criteria for a worklog search.
type SearchFilters struct {
	StartDate   string   // "2006-01-02"
	EndDate     string   // "2006-01-02"
	AuthorIDs   []string // Jira accountId values
	ProjectKeys []string // Jira project keys
	TeamIDs     []int    // internal DB team IDs (resolved to AuthorIDs by caller)
}

// WorklogItem is a processed, display-ready worklog entry.
type WorklogItem struct {
	IssueKey         string
	IssueSummary     string
	WorkReference    string // value of the work-ref custom field
	Author           User
	TimeSpentSeconds int
	Started          time.Time
	Comment          string
}

// WorklogGroup groups WorklogItems by WorkReference.
type WorklogGroup struct {
	WorkReference string
	TotalSeconds  int
	Items         []WorklogItem
}

// ProgressEvent carries status from a background search goroutine to the UI.
type ProgressEvent struct {
	Type      string // "searching" | "processing" | "finalizing"
	Pages     int
	Found     int
	Processed int
	Total     int
	Current   string // issue key currently being processed
}

// Project represents a Jira project.
type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// Field represents a Jira custom field.
type Field struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
```

- [ ] **Step 2: Build to verify**

```bash
go build ./jira/...
```
Expected: no errors (no functions yet, just types).

---

## Task 2: Jira HTTP Client

**Files:**
- Create: `jira/client.go`
- Create: `jira/client_test.go`

- [ ] **Step 1: Write the failing tests first**

```go
// jira/client_test.go
package jira_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
)

// newMockServer creates a test HTTP server and a client pointed at it.
func newMockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *jira.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv, jira.NewClientWithBaseURL(srv.URL, "test@example.com", "mytoken")
}

func TestClientGetDecodesJSON(t *testing.T) {
	type Payload struct{ Message string }
	_, client := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Payload{"hello"})
	})

	var out Payload
	if err := client.Get(context.Background(), "/test", nil, &out); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if out.Message != "hello" {
		t.Errorf("got %q, want %q", out.Message, "hello")
	}
}

func TestClientGetSendsQueryParams(t *testing.T) {
	type Payload struct{ Q string }
	_, client := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Payload{r.URL.Query().Get("jql")})
	})

	var out Payload
	err := client.Get(context.Background(), "/search", map[string]string{"jql": "project=ABC"}, &out)
	if err != nil {
		t.Fatalf("Get with params: %v", err)
	}
	if out.Q != "project=ABC" {
		t.Errorf("got %q, want %q", out.Q, "project=ABC")
	}
}

func TestClientGetSendsAuthHeader(t *testing.T) {
	_, client := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("Authorization header missing")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Write([]byte("{}"))
	})
	var out struct{}
	client.Get(context.Background(), "/test", nil, &out)
}

func TestClientReturnsAPIErrorOnNon2xx(t *testing.T) {
	_, client := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"errorMessages":["not found"]}`, http.StatusNotFound)
	})

	var out struct{}
	err := client.Get(context.Background(), "/missing", nil, &out)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*jira.APIError)
	if !ok {
		t.Fatalf("expected *jira.APIError, got %T: %v", err, err)
	}
	if apiErr.Status != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", apiErr.Status)
	}
}

func TestClientPostEncodesBody(t *testing.T) {
	type Request struct{ JQL string `json:"jql"` }
	type Response struct{ Received string }
	_, client := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req Request
		json.NewDecoder(r.Body).Decode(&req)
		json.NewEncoder(w).Encode(Response{req.JQL})
	})

	var out Response
	err := client.Post(context.Background(), "/search", Request{"project=X"}, &out)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if out.Received != "project=X" {
		t.Errorf("got %q, want %q", out.Received, "project=X")
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./jira/... -run TestClient -v
```
Expected: compilation failure — `jira.NewClientWithBaseURL` and `jira.APIError` don't exist yet.

- [ ] **Step 3: Create `jira/client.go`**

```go
// jira/client.go
package jira

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// APIError is returned when Jira responds with a non-2xx HTTP status.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Jira API %d: %s", e.Status, e.Message)
}

// Client is a thin net/http wrapper for the Jira Cloud REST API v3.
type Client struct {
	baseURL string
	headers http.Header
	http    *http.Client
}

// NewClient builds a Client pointed at a real Jira Cloud instance.
// domain may be "yourorg.atlassian.net" or "https://yourorg.atlassian.net".
func NewClient(domain, email, apiToken string) *Client {
	clean := strings.TrimRight(
		strings.TrimPrefix(strings.TrimPrefix(domain, "https://"), "http://"),
		"/",
	)
	return newClient("https://"+clean+"/rest/api/3", email, apiToken)
}

// NewClientWithBaseURL creates a Client with a custom base URL — used in tests.
func NewClientWithBaseURL(baseURL, email, apiToken string) *Client {
	return newClient(strings.TrimRight(baseURL, "/"), email, apiToken)
}

func newClient(baseURL, email, apiToken string) *Client {
	creds := base64.StdEncoding.EncodeToString([]byte(email + ":" + apiToken))
	return &Client{
		baseURL: baseURL,
		headers: http.Header{
			"Authorization": {"Basic " + creds},
			"Accept":        {"application/json"},
			"Content-Type":  {"application/json"},
		},
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

// Get sends GET baseURL+path with optional query params, decoding JSON into out.
func (c *Client) Get(ctx context.Context, path string, params map[string]string, out any) error {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}
	if len(params) > 0 {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header = c.headers.Clone()
	return c.do(req, out)
}

// Post sends POST baseURL+path with a JSON-encoded body, decoding JSON into out.
func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	req.Header = c.headers.Clone()
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{Status: resp.StatusCode, Message: string(body)}
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
```

- [ ] **Step 4: Run the tests — expect PASS**

```bash
go test ./jira/... -run TestClient -v
```
Expected:
```
--- PASS: TestClientGetDecodesJSON
--- PASS: TestClientGetSendsQueryParams
--- PASS: TestClientGetSendsAuthHeader
--- PASS: TestClientReturnsAPIErrorOnNon2xx
--- PASS: TestClientPostEncodesBody
PASS
```

- [ ] **Step 5: Commit**

```bash
git add jira/types.go jira/client.go jira/client_test.go
git commit -m "feat: jira HTTP client with Basic Auth + tests"
```

---

## Task 3: Worklog Fetch Pipeline

**Files:**
- Create: `jira/worklogs.go`
- Create: `jira/worklogs_test.go`
- Create: `jira/helpers.go` (shared internal helpers)

- [ ] **Step 1: Write the failing tests first**

```go
// jira/worklogs_test.go
package jira_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
)

// buildIssueJSON creates a minimal Jira search response JSON for testing.
func buildSearchResponse(issues []map[string]any) []byte {
	resp := map[string]any{
		"issues":        issues,
		"nextPageToken": "",
		"total":         len(issues),
	}
	b, _ := json.Marshal(resp)
	return b
}

func TestFetchWorklogsGroupsByWorkRef(t *testing.T) {
	issues := []map[string]any{
		{
			"key": "JSW-1",
			"fields": map[string]any{
				"summary": "Task A",
				"project": map[string]any{"key": "JSW", "name": "JSW Project"},
				"worklog": map[string]any{
					"total": 1,
					"worklogs": []map[string]any{
						{
							"id":               "wl1",
							"timeSpentSeconds": 3600,
							"started":          "2026-01-15T09:00:00.000+0000",
							"author":           map[string]any{"accountId": "user1", "displayName": "Alice"},
						},
					},
				},
				"customfield_10001": "CR-001",
			},
		},
		{
			"key": "JSW-2",
			"fields": map[string]any{
				"summary": "Task B",
				"project": map[string]any{"key": "JSW", "name": "JSW Project"},
				"worklog": map[string]any{
					"total": 1,
					"worklogs": []map[string]any{
						{
							"id":               "wl2",
							"timeSpentSeconds": 7200,
							"started":          "2026-01-15T10:00:00.000+0000",
							"author":           map[string]any{"accountId": "user1", "displayName": "Alice"},
						},
					},
				},
				"customfield_10001": "CR-001",
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(buildSearchResponse(issues))
	}))
	defer srv.Close()

	client := jira.NewClientWithBaseURL(srv.URL, "test@example.com", "token")
	filters := jira.SearchFilters{StartDate: "2026-01-01", EndDate: "2026-01-31"}
	progress := make(chan jira.ProgressEvent, 20)

	groups, err := jira.FetchWorklogs(context.Background(), client, filters, "customfield_10001", progress)
	if err != nil {
		t.Fatalf("FetchWorklogs: %v", err)
	}

	if len(groups) != 1 {
		t.Fatalf("expected 1 group (CR-001), got %d", len(groups))
	}
	if groups[0].WorkReference != "CR-001" {
		t.Errorf("WorkReference: got %q, want %q", groups[0].WorkReference, "CR-001")
	}
	if groups[0].TotalSeconds != 10800 {
		t.Errorf("TotalSeconds: got %d, want 10800", groups[0].TotalSeconds)
	}
	if len(groups[0].Items) != 2 {
		t.Errorf("Items count: got %d, want 2", len(groups[0].Items))
	}
}

func TestFetchWorklogsRespectsDateFilter(t *testing.T) {
	issues := []map[string]any{
		{
			"key": "JSW-1",
			"fields": map[string]any{
				"summary": "Task",
				"project": map[string]any{"key": "JSW", "name": "JSW"},
				"worklog": map[string]any{
					"total": 2,
					"worklogs": []map[string]any{
						// in range
						{"id": "wl1", "timeSpentSeconds": 3600, "started": "2026-01-15T09:00:00.000+0000",
							"author": map[string]any{"accountId": "u1", "displayName": "Alice"}},
						// out of range
						{"id": "wl2", "timeSpentSeconds": 3600, "started": "2026-02-01T09:00:00.000+0000",
							"author": map[string]any{"accountId": "u1", "displayName": "Alice"}},
					},
				},
				"customfield_10001": "CR-001",
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(buildSearchResponse(issues))
	}))
	defer srv.Close()

	client := jira.NewClientWithBaseURL(srv.URL, "test@example.com", "token")
	filters := jira.SearchFilters{StartDate: "2026-01-01", EndDate: "2026-01-31"}
	progress := make(chan jira.ProgressEvent, 20)

	groups, err := jira.FetchWorklogs(context.Background(), client, filters, "customfield_10001", progress)
	if err != nil {
		t.Fatalf("FetchWorklogs: %v", err)
	}

	if len(groups) == 0 {
		t.Fatal("expected 1 group, got 0")
	}
	if groups[0].TotalSeconds != 3600 {
		t.Errorf("TotalSeconds after date filter: got %d, want 3600", groups[0].TotalSeconds)
	}
}

func TestFetchWorklogsCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// slow server — should be cancelled before responding
		w.Write(buildSearchResponse(nil))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := jira.NewClientWithBaseURL(srv.URL, "test@example.com", "token")
	filters := jira.SearchFilters{StartDate: "2026-01-01", EndDate: "2026-01-31"}
	progress := make(chan jira.ProgressEvent, 20)

	_, err := jira.FetchWorklogs(ctx, client, filters, "customfield_10001", progress)
	if err == nil {
		t.Error("expected error due to cancellation, got nil")
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./jira/... -run TestFetch -v
```
Expected: compilation failure — `jira.FetchWorklogs` doesn't exist yet.

- [ ] **Step 3: Create `jira/helpers.go`**

```go
// jira/helpers.go
package jira

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// decodeWorklogs converts a []any (from JSON map) to []Worklog.
func decodeWorklogs(raw []any) []Worklog {
	if raw == nil {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var worklogs []Worklog
	json.Unmarshal(b, &worklogs)
	return worklogs
}

// parseJiraTime parses Jira's timestamp format into time.Time.
// Jira uses offsets like "+0700" or "+0000".
func parseJiraTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000+0700",
		"2006-01-02T15:04:05.000Z",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparseable Jira time: %q", s)
}

// buildWorklogJQL constructs JQL for the worklog date range search.
func buildWorklogJQL(f SearchFilters) string {
	parts := []string{
		fmt.Sprintf("worklogDate >= %q AND worklogDate <= %q", f.StartDate, f.EndDate),
	}
	if len(f.ProjectKeys) > 0 {
		quoted := make([]string, len(f.ProjectKeys))
		for i, k := range f.ProjectKeys {
			quoted[i] = fmt.Sprintf("%q", k)
		}
		parts = append(parts, fmt.Sprintf("project in (%s)", strings.Join(quoted, ",")))
	}
	if len(f.AuthorIDs) > 0 {
		quoted := make([]string, len(f.AuthorIDs))
		for i, id := range f.AuthorIDs {
			quoted[i] = fmt.Sprintf("%q", id)
		}
		parts = append(parts, fmt.Sprintf("worklogAuthor in (%s)", strings.Join(quoted, ",")))
	}
	return strings.Join(parts, " AND ")
}

// containsStr returns true if slice contains val.
func containsStr(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Create `jira/worklogs.go`**

```go
// jira/worklogs.go
package jira

import (
	"context"
	"fmt"
)

// FetchWorklogs fetches all worklogs matching filters and returns them grouped
// by work-reference custom field value. Call this on a goroutine:
//
//	progress := make(chan jira.ProgressEvent, 10)
//	go func() {
//	    groups, err := jira.FetchWorklogs(ctx, client, filters, fieldID, progress)
//	    // read groups on UI thread via fyne.Do
//	}()
//	for ev := range progress { /* update UI */ }
func FetchWorklogs(
	ctx context.Context,
	client *Client,
	filters SearchFilters,
	workRefFieldID string,
	progress chan<- ProgressEvent,
) ([]WorklogGroup, error) {
	defer close(progress)

	// Phase 1: paginate JQL search to collect all matching issues.
	allIssues, err := searchAllIssues(ctx, client, filters, workRefFieldID, progress)
	if err != nil {
		return nil, err
	}

	// Phase 2: for each issue, collect and filter its worklogs.
	items, err := processIssues(ctx, client, allIssues, filters, workRefFieldID, progress)
	if err != nil {
		return nil, err
	}

	progress <- ProgressEvent{Type: "finalizing"}
	return groupByWorkReference(items), nil
}

// searchAllIssues paginates the Jira search API and collects all issues.
func searchAllIssues(
	ctx context.Context,
	client *Client,
	filters SearchFilters,
	workRefFieldID string,
	progress chan<- ProgressEvent,
) ([]Issue, error) {
	jql := buildWorklogJQL(filters)
	fields := []string{"summary", "project", "worklog", workRefFieldID}

	var allIssues []Issue
	var nextPageToken string
	page := 0

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		body := map[string]any{
			"jql":        jql,
			"fields":     fields,
			"maxResults": 50,
		}
		if nextPageToken != "" {
			body["nextPageToken"] = nextPageToken
		}

		var res SearchResponse
		if err := client.Post(ctx, "/search/jql", body, &res); err != nil {
			return nil, fmt.Errorf("search page %d: %w", page+1, err)
		}
		allIssues = append(allIssues, res.Issues...)
		page++

		progress <- ProgressEvent{
			Type:  "searching",
			Pages: page,
			Found: len(allIssues),
		}

		if res.NextPageToken == "" || len(res.Issues) < 50 {
			break
		}
		nextPageToken = res.NextPageToken
	}
	return allIssues, nil
}

// processIssues iterates issues, paginates worklogs if needed, filters by
// date range and author, and builds WorklogItems.
func processIssues(
	ctx context.Context,
	client *Client,
	issues []Issue,
	filters SearchFilters,
	workRefFieldID string,
	progress chan<- ProgressEvent,
) ([]WorklogItem, error) {
	var items []WorklogItem

	for i, issue := range issues {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		progress <- ProgressEvent{
			Type:      "processing",
			Processed: i + 1,
			Total:     len(issues),
			Current:   issue.Key,
		}

		worklogs := issue.InlineWorklogs()

		// If there are more worklogs than the 20 inline, fetch them all.
		if issue.WorklogTotal() > len(worklogs) {
			all, err := fetchAllWorklogs(ctx, client, issue.Key)
			if err != nil {
				return nil, fmt.Errorf("fetch worklogs for %s: %w", issue.Key, err)
			}
			worklogs = all
		}

		workRef := issue.CustomField(workRefFieldID)

		for _, wl := range worklogs {
			t, err := parseJiraTime(wl.Started)
			if err != nil {
				continue
			}
			dateStr := t.Format("2006-01-02")
			if dateStr < filters.StartDate || dateStr > filters.EndDate {
				continue
			}
			if len(filters.AuthorIDs) > 0 && !containsStr(filters.AuthorIDs, wl.Author.AccountID) {
				continue
			}
			items = append(items, WorklogItem{
				IssueKey:         issue.Key,
				IssueSummary:     issue.Summary(),
				WorkReference:    workRef,
				Author:           wl.Author,
				TimeSpentSeconds: wl.TimeSpentSeconds,
				Started:          t,
				Comment:          wl.Comment,
			})
		}
	}
	return items, nil
}

// fetchAllWorklogs pages through the worklog endpoint for one issue.
func fetchAllWorklogs(ctx context.Context, client *Client, issueKey string) ([]Worklog, error) {
	var all []Worklog
	startAt := 0
	for {
		var resp WorklogListResponse
		if err := client.Get(ctx, fmt.Sprintf("/issue/%s/worklog", issueKey),
			map[string]string{
				"startAt":    fmt.Sprintf("%d", startAt),
				"maxResults": "100",
			}, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Worklogs...)
		startAt += len(resp.Worklogs)
		if startAt >= resp.Total || len(resp.Worklogs) == 0 {
			break
		}
	}
	return all, nil
}

// groupByWorkReference groups WorklogItems by their WorkReference, preserving order.
func groupByWorkReference(items []WorklogItem) []WorklogGroup {
	order := []string{}
	groups := map[string]*WorklogGroup{}

	for _, item := range items {
		key := item.WorkReference
		if key == "" {
			key = "(no work reference)"
		}
		g, ok := groups[key]
		if !ok {
			g = &WorklogGroup{WorkReference: key}
			groups[key] = g
			order = append(order, key)
		}
		g.Items = append(g.Items, item)
		g.TotalSeconds += item.TimeSpentSeconds
	}

	result := make([]WorklogGroup, 0, len(order))
	for _, key := range order {
		result = append(result, *groups[key])
	}
	return result
}
```

- [ ] **Step 5: Run the tests — expect PASS**

```bash
go test ./jira/... -v
```
Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add jira/
git commit -m "feat: jira worklog fetch pipeline with pagination, filtering, progress channel"
```

---

## Task 4: Users, Projects, Fields Fetchers

**Files:**
- Create: `jira/users.go`
- Create: `jira/projects.go`
- Create: `jira/fields.go`

- [ ] **Step 1: Create `jira/users.go`**

```go
// jira/users.go
package jira

import (
	"context"
	"fmt"
)

// FetchUsers retrieves all active Jira users via the user search API.
// Jira Cloud paginates at 50; we loop until exhausted.
func FetchUsers(ctx context.Context, client *Client) ([]User, error) {
	var all []User
	startAt := 0
	for {
		var page []User
		err := client.Get(ctx, "/users/search", map[string]string{
			"startAt":    fmt.Sprintf("%d", startAt),
			"maxResults": "50",
		}, &page)
		if err != nil {
			return nil, fmt.Errorf("fetch users at %d: %w", startAt, err)
		}
		// Filter inactive users
		for _, u := range page {
			if u.Active {
				all = append(all, u)
			}
		}
		startAt += len(page)
		if len(page) < 50 {
			break
		}
	}
	return all, nil
}
```

- [ ] **Step 2: Create `jira/projects.go`**

```go
// jira/projects.go
package jira

import (
	"context"
	"fmt"
)

type projectsResponse struct {
	Values     []Project `json:"values"`
	IsLast     bool      `json:"isLast"`
	StartAt    int       `json:"startAt"`
	MaxResults int       `json:"maxResults"`
}

// FetchProjects retrieves all Jira projects the user can access.
func FetchProjects(ctx context.Context, client *Client) ([]Project, error) {
	var all []Project
	startAt := 0
	for {
		var resp projectsResponse
		err := client.Get(ctx, "/project/search", map[string]string{
			"startAt":    fmt.Sprintf("%d", startAt),
			"maxResults": "50",
		}, &resp)
		if err != nil {
			return nil, fmt.Errorf("fetch projects at %d: %w", startAt, err)
		}
		all = append(all, resp.Values...)
		if resp.IsLast || len(resp.Values) == 0 {
			break
		}
		startAt += len(resp.Values)
	}
	return all, nil
}
```

- [ ] **Step 3: Create `jira/fields.go`**

```go
// jira/fields.go
package jira

import "context"

// FetchFields retrieves all Jira fields (built-in + custom).
// Used in Settings to let the user pick the Work Reference custom field ID.
func FetchFields(ctx context.Context, client *Client) ([]Field, error) {
	var fields []Field
	if err := client.Get(ctx, "/field", nil, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}
```

- [ ] **Step 4: Build to verify**

```bash
go build ./jira/...
```
Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add jira/users.go jira/projects.go jira/fields.go
git commit -m "feat: jira users/projects/fields fetchers"
```

---

## Task 5: Database — Open + Schema Migration

**Files:**
- Create: `db/types.go`
- Create: `db/db.go`

- [ ] **Step 1: Create `db/types.go`**

```go
// db/types.go
package db

import "time"

// AppConfig holds Jira connection settings.
// ApiToken is NOT stored in the DB — it lives in fyne.Preferences (OS keychain).
type AppConfig struct {
	JiraDomain      string
	Email           string
	ApiToken        string // populated by caller from fyne.Preferences
	WorkRefFieldID  string
	VerticalFieldID string
	CompanyFieldID  string
	UatEndFieldID   string
}

// Team is a named group of Jira users.
type Team struct {
	ID   int
	Name string
}

// TeamMember links a Jira user to a Team with optional date range.
type TeamMember struct {
	ID        int
	TeamID    int
	UserID    string // Jira accountId
	JoinDate  string // "2006-01-02" or ""
	LeaveDate string // "2006-01-02" or ""
}

// PublicHoliday is a non-working day used in timesheet calculations.
type PublicHoliday struct {
	ID   int
	Date string // "2006-01-02"
	Name string
}

// JiraUser is a cached Jira user record.
type JiraUser struct {
	ID           string // Jira accountId
	DisplayName  string
	EmailAddress string
	AvatarURL    string
	Active       bool
	SyncedAt     time.Time
}
```

- [ ] **Step 2: Create `db/db.go`**

```go
// db/db.go
package db

import (
	"database/sql"

	_ "modernc.org/sqlite" // pure-Go SQLite driver — no CGO required
)

// Repository is the root of all DB access. Pass it by pointer throughout the app.
type Repository struct {
	Config   *ConfigRepo
	Teams    *TeamRepo
	Holidays *HolidayRepo
	Users    *UserRepo
	db       *sql.DB
}

// Open opens (or creates) the SQLite database at path, runs the schema
// migration, and returns a fully initialised Repository.
// Use ":memory:" for tests.
func Open(path string) *Repository {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		panic("db.Open: " + err.Error())
	}
	conn.SetMaxOpenConns(1) // SQLite supports only one writer at a time
	migrate(conn)
	return &Repository{
		Config:   &ConfigRepo{conn},
		Teams:    &TeamRepo{conn},
		Holidays: &HolidayRepo{conn},
		Users:    &UserRepo{conn},
		db:       conn,
	}
}

// Close closes the underlying database connection.
func (r *Repository) Close() error {
	return r.db.Close()
}

// migrate creates all tables if they don't already exist.
// This is an append-only migration — safe to run on every startup.
func migrate(db *sql.DB) {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS app_config (
			id               INTEGER PRIMARY KEY DEFAULT 1,
			jira_domain      TEXT NOT NULL DEFAULT '',
			email            TEXT NOT NULL DEFAULT '',
			work_ref_field_id  TEXT NOT NULL DEFAULT '',
			vertical_field_id  TEXT NOT NULL DEFAULT '',
			company_field_id   TEXT NOT NULL DEFAULT '',
			uat_end_field_id   TEXT NOT NULL DEFAULT '',
			updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// api_token is intentionally absent — stored in OS keychain via fyne.Preferences

		`CREATE TABLE IF NOT EXISTS teams (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS team_members (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			team_id    INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
			user_id    TEXT NOT NULL,
			join_date  TEXT NOT NULL DEFAULT '',
			leave_date TEXT NOT NULL DEFAULT '',
			UNIQUE(team_id, user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS public_holidays (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS jira_users (
			id            TEXT PRIMARY KEY,
			display_name  TEXT NOT NULL,
			email_address TEXT NOT NULL DEFAULT '',
			avatar_url    TEXT NOT NULL DEFAULT '',
			active        INTEGER NOT NULL DEFAULT 1,
			synced_at     DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			panic("db.migrate: " + err.Error())
		}
	}
}
```

- [ ] **Step 3: Build to verify**

```bash
go build ./db/...
```
Expected: no errors.

---

## Task 6: Config + User Repos + Tests

**Files:**
- Create: `db/config_repo.go`
- Create: `db/user_repo.go`
- Create: `db/db_test.go`

- [ ] **Step 1: Write the failing tests first**

```go
// db/db_test.go
package db_test

import (
	"testing"

	"github.com/uchup07/fyne-jira-worklog-tracker/db"
)

// openTestDB returns an in-memory database — isolated per test call.
func openTestDB(t *testing.T) *db.Repository {
	t.Helper()
	repo := db.Open(":memory:")
	t.Cleanup(func() { repo.Close() })
	return repo
}

// --- Config repo ---

func TestConfigGetEmptyDB(t *testing.T) {
	repo := openTestDB(t)
	cfg, err := repo.Config.Get()
	if err != nil {
		t.Fatalf("Get on empty DB: %v", err)
	}
	if cfg != nil {
		t.Error("expected nil on empty DB")
	}
}

func TestConfigSaveAndGet(t *testing.T) {
	repo := openTestDB(t)
	want := &db.AppConfig{
		JiraDomain:     "myorg.atlassian.net",
		Email:          "me@example.com",
		WorkRefFieldID: "customfield_10001",
	}
	if err := repo.Config.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.Config.Get()
	if err != nil {
		t.Fatalf("Get after Save: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil after save")
	}
	if got.JiraDomain != want.JiraDomain {
		t.Errorf("JiraDomain: got %q, want %q", got.JiraDomain, want.JiraDomain)
	}
	if got.Email != want.Email {
		t.Errorf("Email: got %q, want %q", got.Email, want.Email)
	}
	if got.WorkRefFieldID != want.WorkRefFieldID {
		t.Errorf("WorkRefFieldID: got %q, want %q", got.WorkRefFieldID, want.WorkRefFieldID)
	}
	// ApiToken is NOT stored in DB — should always be empty from Get()
	if got.ApiToken != "" {
		t.Error("ApiToken should not be persisted in DB")
	}
}

func TestConfigSaveUpserts(t *testing.T) {
	repo := openTestDB(t)
	repo.Config.Save(&db.AppConfig{JiraDomain: "old.atlassian.net", Email: "old@e.com"})
	repo.Config.Save(&db.AppConfig{JiraDomain: "new.atlassian.net", Email: "new@e.com"})

	cfg, _ := repo.Config.Get()
	if cfg.JiraDomain != "new.atlassian.net" {
		t.Errorf("upsert failed: got %q", cfg.JiraDomain)
	}
}

// --- User repo ---

func TestUserUpsertAndList(t *testing.T) {
	repo := openTestDB(t)
	users := []db.JiraUser{
		{ID: "u1", DisplayName: "Alice", EmailAddress: "alice@e.com", Active: true},
		{ID: "u2", DisplayName: "Bob", EmailAddress: "bob@e.com", Active: true},
	}
	if err := repo.Users.Upsert(users); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	got, err := repo.Users.ListActive()
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 users, got %d", len(got))
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./db/... -v
```
Expected: compilation failure — repos not implemented yet.

- [ ] **Step 3: Create `db/config_repo.go`**

```go
// db/config_repo.go
package db

import "database/sql"

// ConfigRepo handles CRUD for app_config.
// NOTE: ApiToken is never written to or read from this table.
// The caller is responsible for reading/writing the token via fyne.Preferences.
type ConfigRepo struct{ db *sql.DB }

// Get retrieves the stored config. Returns nil (no error) if not yet configured.
func (r *ConfigRepo) Get() (*AppConfig, error) {
	row := r.db.QueryRow(`
		SELECT jira_domain, email, work_ref_field_id,
		       vertical_field_id, company_field_id, uat_end_field_id
		FROM app_config WHERE id = 1`)
	cfg := &AppConfig{}
	err := row.Scan(
		&cfg.JiraDomain, &cfg.Email,
		&cfg.WorkRefFieldID, &cfg.VerticalFieldID,
		&cfg.CompanyFieldID, &cfg.UatEndFieldID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// Save inserts or updates the config record (upsert on id=1).
// ApiToken in cfg is intentionally ignored — use fyne.Preferences for that.
func (r *ConfigRepo) Save(cfg *AppConfig) error {
	_, err := r.db.Exec(`
		INSERT INTO app_config
		  (id, jira_domain, email, work_ref_field_id,
		   vertical_field_id, company_field_id, uat_end_field_id, updated_at)
		VALUES (1, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
		  jira_domain      = excluded.jira_domain,
		  email            = excluded.email,
		  work_ref_field_id  = excluded.work_ref_field_id,
		  vertical_field_id  = excluded.vertical_field_id,
		  company_field_id   = excluded.company_field_id,
		  uat_end_field_id   = excluded.uat_end_field_id,
		  updated_at       = excluded.updated_at`,
		cfg.JiraDomain, cfg.Email,
		cfg.WorkRefFieldID, cfg.VerticalFieldID,
		cfg.CompanyFieldID, cfg.UatEndFieldID,
	)
	return err
}
```

- [ ] **Step 4: Create `db/user_repo.go`**

```go
// db/user_repo.go
package db

import (
	"database/sql"
	"time"
)

// UserRepo handles caching of Jira user records.
type UserRepo struct{ db *sql.DB }

// Upsert bulk-inserts or updates Jira user records.
func (r *UserRepo) Upsert(users []JiraUser) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO jira_users (id, display_name, email_address, avatar_url, active, synced_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		  display_name  = excluded.display_name,
		  email_address = excluded.email_address,
		  avatar_url    = excluded.avatar_url,
		  active        = excluded.active,
		  synced_at     = excluded.synced_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	for _, u := range users {
		active := 0
		if u.Active {
			active = 1
		}
		if _, err := stmt.Exec(u.ID, u.DisplayName, u.EmailAddress, u.AvatarURL, active, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListActive returns all users marked active.
func (r *UserRepo) ListActive() ([]JiraUser, error) {
	rows, err := r.db.Query(`
		SELECT id, display_name, email_address, avatar_url, active, synced_at
		FROM jira_users WHERE active = 1 ORDER BY display_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []JiraUser
	for rows.Next() {
		var u JiraUser
		var active int
		var syncedAt string
		if err := rows.Scan(&u.ID, &u.DisplayName, &u.EmailAddress, &u.AvatarURL, &active, &syncedAt); err != nil {
			return nil, err
		}
		u.Active = active == 1
		u.SyncedAt, _ = time.Parse("2006-01-02T15:04:05Z", syncedAt)
		users = append(users, u)
	}
	return users, rows.Err()
}
```

- [ ] **Step 5: Run the tests — expect PASS**

```bash
go test ./db/... -run "TestConfig|TestUser" -v
```
Expected: all targeted tests pass.

---

## Task 7: Team + Holiday Repos

**Files:**
- Create: `db/team_repo.go`
- Create: `db/holiday_repo.go`

- [ ] **Step 1: Create `db/team_repo.go`**

```go
// db/team_repo.go
package db

import "database/sql"

// TeamRepo handles CRUD for teams and team_members.
type TeamRepo struct{ db *sql.DB }

func (r *TeamRepo) ListTeams() ([]Team, error) {
	rows, err := r.db.Query(`SELECT id, name FROM teams ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var teams []Team
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

func (r *TeamRepo) CreateTeam(name string) (int, error) {
	res, err := r.db.Exec(`INSERT INTO teams (name) VALUES (?)`, name)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (r *TeamRepo) DeleteTeam(id int) error {
	_, err := r.db.Exec(`DELETE FROM teams WHERE id = ?`, id)
	return err
}

func (r *TeamRepo) ListMembers(teamID int) ([]TeamMember, error) {
	rows, err := r.db.Query(`
		SELECT id, team_id, user_id, join_date, leave_date
		FROM team_members WHERE team_id = ? ORDER BY user_id`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []TeamMember
	for rows.Next() {
		var m TeamMember
		if err := rows.Scan(&m.ID, &m.TeamID, &m.UserID, &m.JoinDate, &m.LeaveDate); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *TeamRepo) AddMember(m TeamMember) error {
	_, err := r.db.Exec(`
		INSERT INTO team_members (team_id, user_id, join_date, leave_date)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(team_id, user_id) DO UPDATE SET
		  join_date  = excluded.join_date,
		  leave_date = excluded.leave_date`,
		m.TeamID, m.UserID, m.JoinDate, m.LeaveDate)
	return err
}

func (r *TeamRepo) RemoveMember(id int) error {
	_, err := r.db.Exec(`DELETE FROM team_members WHERE id = ?`, id)
	return err
}
```

- [ ] **Step 2: Create `db/holiday_repo.go`**

```go
// db/holiday_repo.go
package db

import "database/sql"

// HolidayRepo handles CRUD for public_holidays.
type HolidayRepo struct{ db *sql.DB }

func (r *HolidayRepo) List() ([]PublicHoliday, error) {
	rows, err := r.db.Query(`SELECT id, date, name FROM public_holidays ORDER BY date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var holidays []PublicHoliday
	for rows.Next() {
		var h PublicHoliday
		if err := rows.Scan(&h.ID, &h.Date, &h.Name); err != nil {
			return nil, err
		}
		holidays = append(holidays, h)
	}
	return holidays, rows.Err()
}

func (r *HolidayRepo) Add(h PublicHoliday) error {
	_, err := r.db.Exec(`INSERT INTO public_holidays (date, name) VALUES (?, ?)
		ON CONFLICT(date) DO UPDATE SET name = excluded.name`,
		h.Date, h.Name)
	return err
}

func (r *HolidayRepo) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM public_holidays WHERE id = ?`, id)
	return err
}
```

- [ ] **Step 3: Add team + holiday tests to `db/db_test.go`**

```go
// Append to db/db_test.go:

func TestTeamCreateAndList(t *testing.T) {
	repo := openTestDB(t)

	id, err := repo.Teams.CreateTeam("Backend")
	if err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero ID")
	}

	teams, err := repo.Teams.ListTeams()
	if err != nil {
		t.Fatalf("ListTeams: %v", err)
	}
	if len(teams) != 1 || teams[0].Name != "Backend" {
		t.Errorf("unexpected teams: %+v", teams)
	}
}

func TestTeamMemberAddAndList(t *testing.T) {
	repo := openTestDB(t)
	teamID, _ := repo.Teams.CreateTeam("Frontend")

	err := repo.Teams.AddMember(db.TeamMember{
		TeamID:   teamID,
		UserID:   "user-abc",
		JoinDate: "2026-01-01",
	})
	if err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	members, err := repo.Teams.ListMembers(teamID)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	if len(members) != 1 || members[0].UserID != "user-abc" {
		t.Errorf("unexpected members: %+v", members)
	}
}

func TestHolidayAddAndList(t *testing.T) {
	repo := openTestDB(t)

	err := repo.Holidays.Add(db.PublicHoliday{Date: "2026-08-17", Name: "Independence Day"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	holidays, err := repo.Holidays.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(holidays) != 1 || holidays[0].Name != "Independence Day" {
		t.Errorf("unexpected holidays: %+v", holidays)
	}
}
```

- [ ] **Step 4: Run all DB tests**

```bash
go test ./db/... -v
```
Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
git add db/
git commit -m "feat: SQLite DB layer — config, teams, holidays, user cache + tests"
```

---

## Task 8: Wire DB into App + Phase 2 Checkpoint

**Files:**
- Modify: `app/app.go`

- [ ] **Step 1: Update `app/app.go` to hold the DB repo and a Jira client builder**

```go
// app/app.go
package app

import (
	"github.com/uchup07/fyne-jira-worklog-tracker/custom"
	"github.com/uchup07/fyne-jira-worklog-tracker/db"
	"github.com/uchup07/fyne-jira-worklog-tracker/jira"
	"github.com/uchup07/fyne-jira-worklog-tracker/state"

	"fyne.io/fyne/v2"
)

// App is the root of the application.
type App struct {
	fyneApp fyne.App
	window  fyne.Window

	filterState  *state.FilterState
	worklogState *state.WorklogState
	reportState  *state.ReportState

	repo *db.Repository
}

// New initialises the App. DB is opened from Fyne's app-storage directory
// so it persists between runs on the user's machine.
func New(a fyne.App, w fyne.Window) *App {
	a.Settings().SetTheme(custom.NewAppTheme())

	dbPath := a.Storage().RootURI().Path() + "/worklog.db"
	repo := db.Open(dbPath)

	return &App{
		fyneApp:      a,
		window:       w,
		filterState:  state.NewFilterState(),
		worklogState: state.NewWorklogState(),
		reportState:  state.NewReportState(),
		repo:         repo,
	}
}

// Run wires navigation and starts the Fyne event loop.
func (a *App) Run() {
	a.window.SetContent(a.buildNav())
	a.window.ShowAndRun()
}

// JiraClient builds a Jira client from the stored config + the OS-keychain token.
// Returns nil if config is not yet saved.
func (a *App) JiraClient() *jira.Client {
	cfg, err := a.repo.Config.Get()
	if err != nil || cfg == nil {
		return nil
	}
	cfg.ApiToken = a.fyneApp.Preferences().String("api_token")
	if cfg.ApiToken == "" {
		return nil
	}
	return jira.NewClient(cfg.JiraDomain, cfg.Email, cfg.ApiToken)
}
```

- [ ] **Step 2: Full build**

```bash
go build ./...
```
Expected: no errors.

- [ ] **Step 3: Run all tests**

```bash
go test ./...
```
Expected: all tests pass.

- [ ] **Step 4: Run the app and confirm it still opens**

```bash
go run .
```
Expected: same window as Plan 1. No visible change — DB is wired silently.

- [ ] **Step 5: Commit and push**

```bash
git add app/app.go
git commit -m "feat: wire SQLite DB and Jira client builder into App struct"
git push origin main
```

---

## Phase 2 Complete ✓

**What you have:**
- Full Jira HTTP client with Basic Auth, pagination, error handling, and tests
- Worklog fetch pipeline: goroutine-friendly, progress channel, cancellation via context
- Users/Projects/Fields fetchers
- SQLite DB layer: AppConfig, Teams, Members, Holidays, UserCache — all with tests
- `app.App` holds the DB repo and can build a Jira client from stored credentials

**Next:** Plan 3 — Core Screens (Dashboard + Report fully wired with real data)
