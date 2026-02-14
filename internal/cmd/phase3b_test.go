package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExecuteMeetingsListJSON(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/meetings" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":       []map[string]any{{"id": map[string]any{"meeting_id": "m1"}, "title": "Pipeline Review"}},
			"pagination": map[string]any{"next_cursor": ""},
		})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "meetings", "list", "--limit", "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, `"title": "Pipeline Review"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestExecuteAttributesListJSON(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/objects/people/attributes" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{"id": map[string]any{"attribute_id": "a1"}, "title": "Stage", "api_type": "status"}},
		})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "attributes", "list", "objects", "people"})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, `"title": "Stage"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}
