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
