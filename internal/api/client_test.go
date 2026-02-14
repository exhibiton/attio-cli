package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientDoSuccessAndHeaders(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/test" {
			t.Fatalf("expected path /v2/test, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("expected auth header, got %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected content-type application/json, got %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("expected accept application/json, got %q", got)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["hello"] != "world" {
			t.Fatalf("unexpected request body: %#v", body)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client := NewClient("test-key", srv.URL)
	var out map[string]any
	err := client.do(context.Background(), http.MethodPost, "/v2/test", map[string]any{"hello": "world"}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["ok"] != true {
		t.Fatalf("expected ok=true, got %#v", out)
	}
}

func TestClientDoNoContent(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient("test-key", srv.URL)
	if err := client.do(context.Background(), http.MethodDelete, "/v2/test", nil, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestClientDoAPIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status_code":401,"type":"auth_error","code":"unauthorized","message":"bad key"}`))
	}))
	defer srv.Close()

	client := NewClient("test-key", srv.URL)
	err := client.do(context.Background(), http.MethodGet, "/v2/test", nil, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
	attioErr, ok := err.(*AttioError)
	if !ok {
		t.Fatalf("expected *AttioError, got %T", err)
	}
	if attioErr.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", attioErr.StatusCode)
	}
}

func TestClientGetSelf(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/self" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"active":true,"workspace_name":"Test Workspace","workspace_slug":"test-workspace"}`))
	}))
	defer srv.Close()

	client := NewClient("test-key", strings.TrimRight(srv.URL, "/")+"/")
	self, err := client.GetSelf(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !self.Active {
		t.Fatalf("expected active=true")
	}
	if self.WorkspaceName != "Test Workspace" {
		t.Fatalf("unexpected workspace name: %q", self.WorkspaceName)
	}
}

func TestClientUserAgentAndTimeoutSetters(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "attio-cli/test-suite" {
			t.Fatalf("expected custom user-agent, got %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Fatalf("expected empty auth header for empty api key, got %q", got)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client := NewClient("", srv.URL)
	client.SetUserAgent("attio-cli/test-suite")
	client.SetTimeout(5 * time.Second)

	if client.httpClient.Timeout != 5*time.Second {
		t.Fatalf("expected timeout 5s, got %v", client.httpClient.Timeout)
	}

	// Non-positive timeouts should not override the configured timeout.
	client.SetTimeout(0)
	if client.httpClient.Timeout != 5*time.Second {
		t.Fatalf("expected timeout to remain 5s, got %v", client.httpClient.Timeout)
	}

	var out map[string]any
	if err := client.do(context.Background(), http.MethodGet, "/v2/test", nil, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["ok"] != true {
		t.Fatalf("expected ok=true, got %#v", out)
	}
}
