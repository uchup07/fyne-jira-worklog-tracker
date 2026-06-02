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
	type Request struct {
		JQL string `json:"jql"`
	}
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
