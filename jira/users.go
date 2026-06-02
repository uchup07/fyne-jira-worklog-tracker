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
