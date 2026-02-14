package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExecuteIDOnlyMatrix(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"object_id": "o1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/lists/prospects":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"list_id": "l1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/lists/prospects/entries/e1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"entry_id": "e1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/notes/n1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"note_id": "n1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/t1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"task_id": "t1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/comments/c1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"comment_id": "c1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/threads/th1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"thread_id": "th1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/workspace_members/m1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"workspace_member_id": "m1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/webhooks/w1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"webhook_id": "w1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings/m1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"meeting_id": "m1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings/m1/call_recordings/r1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"call_recording_id": "r1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/records/r1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"record_id": "r1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/attributes/a1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"attribute_id": "a1"}}})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/objects/people/attributes/a1/options/o1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"option_id": "o1"}}})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/objects/people/attributes/a1/statuses/s1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"status_id": "s1"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	cases := []struct {
		name string
		args []string
		want string
	}{
		{name: "objects-get", args: []string{"--id-only", "objects", "get", "people"}, want: "o1"},
		{name: "lists-get", args: []string{"--id-only", "lists", "get", "prospects"}, want: "l1"},
		{name: "entries-get", args: []string{"--id-only", "entries", "get", "prospects", "e1"}, want: "e1"},
		{name: "notes-get", args: []string{"--id-only", "notes", "get", "n1"}, want: "n1"},
		{name: "tasks-get", args: []string{"--id-only", "tasks", "get", "t1"}, want: "t1"},
		{name: "comments-get", args: []string{"--id-only", "comments", "get", "c1"}, want: "c1"},
		{name: "threads-get", args: []string{"--id-only", "threads", "get", "th1"}, want: "th1"},
		{name: "members-get", args: []string{"--id-only", "members", "get", "m1"}, want: "m1"},
		{name: "webhooks-get", args: []string{"--id-only", "webhooks", "get", "w1"}, want: "w1"},
		{name: "meetings-get", args: []string{"--id-only", "meetings", "get", "m1"}, want: "m1"},
		{name: "recordings-get", args: []string{"--id-only", "meetings", "recordings", "get", "m1", "r1"}, want: "r1"},
		{name: "records-get", args: []string{"--id-only", "records", "get", "people", "r1"}, want: "r1"},
		{name: "attributes-get", args: []string{"--id-only", "attributes", "get", "objects", "people", "a1"}, want: "a1"},
		{name: "options-update", args: []string{"--id-only", "attributes", "options", "update", "objects", "people", "a1", "o1", "--data", `{"title":"Hot"}`}, want: "o1"},
		{name: "statuses-update", args: []string{"--id-only", "attributes", "statuses", "update", "objects", "people", "a1", "s1", "--data", `{"title":"Open"}`}, want: "s1"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureExecute(t, tc.args)
			if err != nil {
				t.Fatalf("unexpected error: %v stderr=%s", err, stderr)
			}
			if strings.TrimSpace(stdout) != tc.want {
				t.Fatalf("expected id-only output %q, got %q", tc.want, stdout)
			}
		})
	}
}

func TestExecuteAdditionalRunCoverage(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2/lists/prospects/entries/query":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"entry_id": "e1"}}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/people/records/query":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"record_id": "r1"}}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/threads":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"thread_id": "th1"}}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/webhooks":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"webhook_id": "w1"}}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings":
			if r.URL.Query().Get("cursor") == "" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data":       []map[string]any{{"id": map[string]any{"meeting_id": "m1"}}},
					"pagination": map[string]any{"next_cursor": "next"},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}, "pagination": map[string]any{"next_cursor": ""}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":            map[string]any{"object_id": "o1"},
					"api_slug":      "accounts",
					"singular_noun": "Account",
					"plural_noun":   "Accounts",
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/comments/c1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/lists/prospects/entries/e1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/notes/n1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/objects/people/records/r1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/tasks/t1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/webhooks/w1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/meetings/m1/call_recordings/r1":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	jsonCases := []struct {
		name string
		args []string
	}{
		{name: "entries-query-json", args: []string{"--json", "entries", "query", "prospects", "--limit", "1", "--offset", "2"}},
		{name: "records-query-json", args: []string{"--json", "records", "query", "people", "--limit", "1", "--offset", "3"}},
		{name: "threads-list-json", args: []string{"--json", "threads", "list", "--record-id", "rec-1", "--object", "people", "--limit", "1", "--offset", "4"}},
		{name: "webhooks-list-json", args: []string{"--json", "webhooks", "list", "--limit", "1", "--offset", "5"}},
		{name: "objects-create", args: []string{"--json", "objects", "create", "--data", `{"api_slug":"accounts"}`}},
	}

	for _, tc := range jsonCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureExecute(t, tc.args)
			if err != nil {
				t.Fatalf("unexpected error: %v stderr=%s", err, stderr)
			}
			if tc.name == "objects-create" {
				if !strings.Contains(stdout, `"object_id": "o1"`) {
					t.Fatalf("unexpected objects create output: %s", stdout)
				}
				return
			}
			if !strings.Contains(stdout, `"pagination"`) {
				t.Fatalf("expected pagination metadata in output: %s", stdout)
			}
		})
	}

	stdout, stderr, err := captureExecute(t, []string{"meetings", "list", "--all", "--limit", "1", "--max-pages", "2"})
	if err != nil {
		t.Fatalf("meetings list --all failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "ID") || !strings.Contains(stdout, "m1") {
		t.Fatalf("unexpected meetings list --all output: %s", stdout)
	}

	deleteCases := []struct {
		name string
		args []string
		want string
	}{
		{name: "comments-delete", args: []string{"comments", "delete", "c1"}, want: "Deleted comment c1"},
		{name: "entries-delete", args: []string{"entries", "delete", "prospects", "e1"}, want: "Deleted entry e1"},
		{name: "notes-delete", args: []string{"notes", "delete", "n1"}, want: "Deleted note n1"},
		{name: "records-delete", args: []string{"records", "delete", "people", "r1"}, want: "Deleted record r1"},
		{name: "tasks-delete", args: []string{"tasks", "delete", "t1"}, want: "Deleted task t1"},
		{name: "webhooks-delete", args: []string{"webhooks", "delete", "w1"}, want: "Deleted webhook w1"},
		{name: "recordings-delete", args: []string{"meetings", "recordings", "delete", "m1", "r1"}, want: "Deleted call recording r1"},
	}

	for _, tc := range deleteCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := captureExecute(t, tc.args)
			if err != nil {
				t.Fatalf("unexpected delete error: %v stderr=%s", err, stderr)
			}
			if !strings.Contains(stdout, tc.want) {
				t.Fatalf("expected %q in delete output, got %s", tc.want, stdout)
			}
		})
	}
}
