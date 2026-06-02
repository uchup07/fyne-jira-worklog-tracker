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
