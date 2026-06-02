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

// buildSearchResponse creates a minimal Jira search response JSON for testing.
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
