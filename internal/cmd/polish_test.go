package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestExecuteFailEmpty(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/objects" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	_, _, err := captureExecute(t, []string{"--json", "--fail-empty", "objects", "list"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if ExitCode(err) != ExitCodeNoResult {
		t.Fatalf("expected exit code %d, got %d", ExitCodeNoResult, ExitCode(err))
	}
}

func TestExecuteDryRunSkipsNetwork(t *testing.T) {
	setupCLIEnv(t)

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		http.NotFound(w, r)
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "--dry-run", "objects", "create", "--data", `{"api_slug":"new_object"}`})
	if err != nil {
		t.Fatalf("unexpected error: %v stderr=%s", err, stderr)
	}
	if atomic.LoadInt32(&hits) != 0 {
		t.Fatalf("expected no network calls in dry-run, got %d", hits)
	}
	if !strings.Contains(stdout, `"dry_run": true`) {
		t.Fatalf("expected dry-run output, got %s", stdout)
	}
}

func TestExecuteSearchAlias(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/objects/records/search" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"record_text": "Ada Lovelace"}}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "search", "ada", "--objects", "people"})
	if err != nil {
		t.Fatalf("unexpected error: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "Ada Lovelace") {
		t.Fatalf("expected search result in output, got %s", stdout)
	}
}

func TestExecuteQueryAlias(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/objects/people/records/query" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"record_id": "r1"}}}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "query", "people", "--limit", "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"record_id": "r1"`) {
		t.Fatalf("expected query output, got %s", stdout)
	}
}

func TestExecuteCompletionScript(t *testing.T) {
	setupCLIEnv(t)

	stdout, stderr, err := captureExecute(t, []string{"completion", "bash"})
	if err != nil {
		t.Fatalf("unexpected error: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "_attio_complete") {
		t.Fatalf("unexpected completion output: %s", stdout)
	}
}
