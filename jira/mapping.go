// jira/mapping.go
package jira

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// MappingConfig holds the custom field IDs needed for the mapping report.
type MappingConfig struct {
	WorkRefFieldID  string
	VerticalFieldID string
	CompanyFieldID  string
	UatEndFieldID   string
}

// BuildMappingReport fetches JSW issues with a non-empty Work Reference field,
// joins them to IRQ metadata, and returns a MappingReport.
// The reference implementation is lib/jira/mapping.ts in the Next.js app.
func BuildMappingReport(
	ctx context.Context,
	client *Client,
	filters SearchFilters,
	cfg MappingConfig,
	progress chan<- ProgressEvent,
) (*MappingReport, error) {
	defer close(progress)

	jql := fmt.Sprintf(
		`worklogDate >= %q AND worklogDate <= %q AND "%s" is not EMPTY`,
		filters.StartDate, filters.EndDate, cfg.WorkRefFieldID,
	)
	fields := []string{"summary", "project", "worklog", cfg.WorkRefFieldID}

	var allIssues []Issue
	var nextPageToken string
	page := 0

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		body := map[string]any{"jql": jql, "fields": fields, "maxResults": 50}
		if nextPageToken != "" {
			body["nextPageToken"] = nextPageToken
		}
		var res SearchResponse
		if err := client.Post(ctx, "/search/jql", body, &res); err != nil {
			return nil, fmt.Errorf("mapping search page %d: %w", page+1, err)
		}
		allIssues = append(allIssues, res.Issues...)
		page++
		progress <- ProgressEvent{Type: "searching", Pages: page, Found: len(allIssues)}
		if res.NextPageToken == "" || len(res.Issues) < 50 {
			break
		}
		nextPageToken = res.NextPageToken
	}

	var rows []MappingRow
	for i, issue := range allIssues {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		progress <- ProgressEvent{Type: "processing", Processed: i + 1, Total: len(allIssues), Current: issue.Key}

		worklogs := issue.InlineWorklogs()
		if issue.WorklogTotal() > len(worklogs) {
			all, err := fetchAllWorklogs(ctx, client, issue.Key)
			if err != nil {
				return nil, fmt.Errorf("worklogs for %s: %w", issue.Key, err)
			}
			worklogs = all
		}

		totalSecs := 0
		months := map[string]int{}
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
			totalSecs += wl.TimeSpentSeconds
			months[t.Format("2006-01")] += wl.TimeSpentSeconds
		}
		if totalSecs == 0 {
			continue
		}

		workRef := issue.CustomField(cfg.WorkRefFieldID)
		rows = append(rows, MappingRow{
			JSWKey:        issue.Key,
			JSWSummary:    issue.Summary(),
			JSWProjectKey: issue.ProjectKey(),
			TaskName:      parseTaskName(issue.Summary()),
			IRQKey:        workRef,
			TotalSeconds:  totalSecs,
			WorklogMonths: months,
		})
	}

	progress <- ProgressEvent{Type: "finalizing"}

	if err := enrichWithIRQData(ctx, client, rows, cfg); err != nil {
		return nil, err
	}

	return buildMappingReport(rows), nil
}

// enrichWithIRQData fetches IRQ issue metadata in batches of 25.
func enrichWithIRQData(ctx context.Context, client *Client, rows []MappingRow, cfg MappingConfig) error {
	seen := map[string]bool{}
	var keys []string
	for _, r := range rows {
		if r.IRQKey != "" && !seen[r.IRQKey] {
			seen[r.IRQKey] = true
			keys = append(keys, r.IRQKey)
		}
	}
	if len(keys) == 0 {
		return nil
	}

	irqMap := map[string]Issue{}
	const batchSize = 25
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}
		batch := keys[i:end]
		quoted := make([]string, len(batch))
		for j, k := range batch {
			quoted[j] = fmt.Sprintf("%q", k)
		}
		jql := fmt.Sprintf("issue in (%s)", strings.Join(quoted, ","))
		fields := []string{"summary", "status", cfg.VerticalFieldID, cfg.CompanyFieldID, cfg.UatEndFieldID}
		body := map[string]any{"jql": jql, "fields": fields, "maxResults": 50}
		var res SearchResponse
		if err := client.Post(ctx, "/search/jql", body, &res); err != nil {
			return fmt.Errorf("fetch IRQ batch: %w", err)
		}
		for _, issue := range res.Issues {
			irqMap[issue.Key] = issue
		}
	}

	for i := range rows {
		irq, ok := irqMap[rows[i].IRQKey]
		if !ok {
			continue
		}
		rows[i].IRQSummary = irq.Summary()
		if status, ok := irq.Fields["status"].(map[string]any); ok {
			rows[i].IRQStatus, _ = status["name"].(string)
		}
		rows[i].Vertical = irq.CustomField(cfg.VerticalFieldID)
		rows[i].Company = irq.CustomField(cfg.CompanyFieldID)
		rows[i].UATEndDate = irq.CustomField(cfg.UatEndFieldID)
	}
	return nil
}

func buildMappingReport(rows []MappingRow) *MappingReport {
	totalSecs := 0
	irqs := map[string]bool{}
	verticals := map[string]bool{}
	companies := map[string]bool{}
	for _, r := range rows {
		totalSecs += r.TotalSeconds
		irqs[r.IRQKey] = true
		verticals[r.Vertical] = true
		companies[r.Company] = true
	}
	return &MappingReport{
		Rows:            rows,
		TotalCRSeconds:  totalSecs,
		UniqueIRQCount:  len(irqs),
		UniqueVerticals: len(verticals),
		UniqueCompanies: len(companies),
		GeneratedAt:     time.Now(),
	}
}

// AggregateByIRQ collapses flat MappingRows into per-IRQ groups, de-duplicating JSW tasks.
func AggregateByIRQ(rows []MappingRow) []IrqGroup {
	order := []string{}
	groups := map[string]*IrqGroup{}

	for _, r := range rows {
		g, ok := groups[r.IRQKey]
		if !ok {
			g = &IrqGroup{
				IRQKey:     r.IRQKey,
				IRQSummary: r.IRQSummary,
				IRQStatus:  r.IRQStatus,
				Company:    r.Company,
				Vertical:   r.Vertical,
				UATEndDate: r.UATEndDate,
			}
			groups[r.IRQKey] = g
			order = append(order, r.IRQKey)
		}
		duplicate := false
		for _, t := range g.JSWTasks {
			if t.Key == r.JSWKey {
				duplicate = true
				break
			}
		}
		if !duplicate {
			g.JSWTasks = append(g.JSWTasks, IrqJSWTask{
				Key:          r.JSWKey,
				Summary:      r.JSWSummary,
				ProjectKey:   r.JSWProjectKey,
				TaskName:     r.TaskName,
				TotalSeconds: r.TotalSeconds,
			})
		}
		g.TotalSeconds += r.TotalSeconds
	}

	result := make([]IrqGroup, 0, len(order))
	for _, key := range order {
		result = append(result, *groups[key])
	}
	return result
}

// VerticalChartData returns hours per vertical, sorted descending by value.
func VerticalChartData(rows []MappingRow) ChartData {
	totals := map[string]float64{}
	for _, r := range rows {
		v := r.Vertical
		if v == "" {
			v = "(unknown)"
		}
		totals[v] += float64(r.TotalSeconds) / 3600
	}
	type kv struct {
		k string
		v float64
	}
	var pairs []kv
	for k, v := range totals {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].v > pairs[j].v })

	cd := ChartData{}
	for _, p := range pairs {
		cd.Labels = append(cd.Labels, p.k)
		cd.Values = append(cd.Values, p.v)
	}
	return cd
}

// parseTaskName extracts a work category from the issue summary.
func parseTaskName(summary string) string {
	lower := strings.ToLower(summary)
	for _, cat := range []string{"development", "testing", "bug fix", "code review", "deployment", "documentation", "analysis"} {
		if strings.Contains(lower, cat) {
			if len(cat) > 0 {
				return strings.ToUpper(cat[:1]) + cat[1:]
			}
		}
	}
	return "Other"
}
