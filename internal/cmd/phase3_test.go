package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExecuteNotesListJSON(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/notes" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("parent_object") != "people" {
			t.Fatalf("expected parent_object query")
		}
		if r.URL.Query().Get("parent_record_id") != "r1" {
			t.Fatalf("expected parent_record_id query")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"note_id": "n1"}, "title": "Test Note"}}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "notes", "list", "--parent-object", "people", "--parent-record", "r1"})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, `"title": "Test Note"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestExecuteMembersListJSON(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/workspace_members" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{"id": map[string]any{"workspace_member_id": "m1"}, "email_address": "m1@example.com", "name": "Member One"}},
		})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "members", "list"})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, `"email_address": "m1@example.com"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}
