package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExecuteObjectsListJSON(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/objects" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{"id": map[string]any{"object_id": "o1"}, "api_slug": "people", "singular_noun": "Person", "plural_noun": "People"}},
		})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "objects", "list"})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, `"api_slug": "people"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestExecuteRecordsQueryAllJSON(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/objects/people/records/query" {
			http.NotFound(w, r)
			return
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		offset, _ := body["offset"].(float64)
		if offset == 0 {
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"record_id": "r1"}}, {"id": map[string]any{"record_id": "r2"}}}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"record_id": "r3"}}}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "records", "query", "people", "--limit", "2", "--all"})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if bytes.Count([]byte(stdout), []byte(`"record_id"`)) != 3 {
		t.Fatalf("expected 3 records in output, got: %s", stdout)
	}
}
