package cmd

import (
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func TestGoldenObjectsListOutputs(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/objects" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":       map[string]any{"object_id": "o1", "workspace_id": "w1"},
					"api_slug": "people", "singular_noun": "Person", "plural_noun": "People",
				},
				{
					"id":       map[string]any{"object_id": "o2", "workspace_id": "w1"},
					"api_slug": "companies", "singular_noun": "Company", "plural_noun": "Companies",
				},
			},
		})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	tests := []struct {
		name   string
		args   []string
		golden string
	}{
		{name: "table", args: []string{"objects", "list"}, golden: "objects_list_table.golden"},
		{name: "json", args: []string{"--json", "objects", "list"}, golden: "objects_list_json.golden"},
		{name: "plain", args: []string{"--plain", "objects", "list"}, golden: "objects_list_plain.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := captureExecute(t, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v stderr=%s", err, stderr)
			}
			if strings.TrimSpace(stderr) != "" {
				t.Fatalf("expected empty stderr, got %q", stderr)
			}
			assertGolden(t, tt.golden, normalizeGolden(stdout))
		})
	}
}

func TestGoldenRecordsQueryOutputs(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/objects/people/records/query" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":         map[string]any{"record_id": "r1"},
					"created_at": "2025-01-01T00:00:00Z",
					"web_url":    "https://app.attio.test/r/1",
					"values": map[string]any{
						"name":            []any{map[string]any{"full_name": "Ada Lovelace"}},
						"email_addresses": []any{map[string]any{"email_address": "ada@example.com"}},
					},
				},
			},
		})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	tests := []struct {
		name   string
		args   []string
		golden string
	}{
		{name: "table", args: []string{"records", "query", "people", "--limit", "1"}, golden: "records_query_table.golden"},
		{name: "json", args: []string{"--json", "records", "query", "people", "--limit", "1"}, golden: "records_query_json.golden"},
		{name: "plain", args: []string{"--plain", "records", "query", "people", "--limit", "1"}, golden: "records_query_plain.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := captureExecute(t, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v stderr=%s", err, stderr)
			}
			if strings.TrimSpace(stderr) != "" {
				t.Fatalf("expected empty stderr, got %q", stderr)
			}
			assertGolden(t, tt.golden, normalizeGolden(stdout))
		})
	}
}

func TestGoldenTasksListOutputs(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/tasks" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":           map[string]any{"task_id": "t1"},
					"content":      "Follow up with prospect",
					"is_completed": false,
					"deadline_at":  "2025-01-05T12:00:00Z",
					"assignees": []any{
						map[string]any{"workspace_member_email_address": "owner@example.com"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	tests := []struct {
		name   string
		args   []string
		golden string
	}{
		{name: "table", args: []string{"tasks", "list", "--limit", "1"}, golden: "tasks_list_table.golden"},
		{name: "json", args: []string{"--json", "tasks", "list", "--limit", "1"}, golden: "tasks_list_json.golden"},
		{name: "plain", args: []string{"--plain", "tasks", "list", "--limit", "1"}, golden: "tasks_list_plain.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := captureExecute(t, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v stderr=%s", err, stderr)
			}
			if strings.TrimSpace(stderr) != "" {
				t.Fatalf("expected empty stderr, got %q", stderr)
			}
			assertGolden(t, tt.golden, normalizeGolden(stdout))
		})
	}
}

func TestGoldenMeetingsListOutputs(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/meetings" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":                map[string]any{"meeting_id": "m1"},
					"title":             "Pipeline Review",
					"start_at":          "2025-01-10T10:00:00Z",
					"end_at":            "2025-01-10T10:30:00Z",
					"participant_count": 3,
				},
			},
			"pagination": map[string]any{"next_cursor": "next_1"},
		})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	tests := []struct {
		name   string
		args   []string
		golden string
	}{
		{name: "table", args: []string{"meetings", "list", "--limit", "1"}, golden: "meetings_list_table.golden"},
		{name: "json", args: []string{"--json", "meetings", "list", "--limit", "1"}, golden: "meetings_list_json.golden"},
		{name: "plain", args: []string{"--plain", "meetings", "list", "--limit", "1"}, golden: "meetings_list_plain.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := captureExecute(t, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v stderr=%s", err, stderr)
			}
			if strings.TrimSpace(stderr) != "" {
				t.Fatalf("expected empty stderr, got %q", stderr)
			}
			assertGolden(t, tt.golden, normalizeGolden(stdout))
		})
	}
}

func assertGolden(t *testing.T, file string, got string) {
	t.Helper()
	path := filepath.Join("testdata", file)

	if *updateGolden {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden file: %v", err)
		}
		return
	}

	wantBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	want := string(wantBytes)
	if want != got {
		t.Fatalf("golden mismatch for %s\nwant:\n%s\n\ngot:\n%s", file, want, got)
	}
}

func normalizeGolden(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return s
}
