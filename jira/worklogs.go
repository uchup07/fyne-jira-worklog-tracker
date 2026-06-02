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
//	    // update UI via fyne.Do after channel closes
//	}()
//	for ev := range progress { /* update progress bar */ }
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

// searchAllIssues paginates the Jira search API and collects all matching issues.
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

// processIssues iterates issues, paginates worklogs if needed, filters by date
// range and author, and returns flat WorklogItems.
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

		// If there are more worklogs than the inline batch, fetch them all.
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

// groupByWorkReference groups WorklogItems by their WorkReference, preserving insertion order.
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
