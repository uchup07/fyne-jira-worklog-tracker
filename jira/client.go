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
