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
