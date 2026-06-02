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
