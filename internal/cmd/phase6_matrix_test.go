package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestExecuteObjectsAndRecordsCommands(t *testing.T) {
	setupCLIEnv(t)

	object := map[string]any{
		"id":            map[string]any{"object_id": "o1"},
		"api_slug":      "people",
		"singular_noun": "Person",
		"plural_noun":   "People",
	}
	record := map[string]any{
		"id":         map[string]any{"record_id": "r1"},
		"created_at": "2025-01-01T00:00:00Z",
		"web_url":    "https://app.attio.com/r/1",
		"values": map[string]any{
			"name":            []any{map[string]any{"full_name": "Ada Lovelace"}},
			"email_addresses": []any{map[string]any{"email_address": "ada@example.com"}},
		},
	}
	entry := map[string]any{
		"id":         map[string]any{"entry_id": "e1"},
		"created_at": "2025-01-02T00:00:00Z",
		"web_url":    "https://app.attio.com/e/1",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": object})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/objects/people":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":            map[string]any{"object_id": "o1"},
					"api_slug":      "people",
					"singular_noun": "Person Updated",
					"plural_noun":   "People",
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/people/records":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": record})
		case r.Method == http.MethodPut && r.URL.Path == "/v2/objects/people/records":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": record})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/people/records/query":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			offset, _ := body["offset"].(float64)
			if offset == 0 {
				_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{record}})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/records/search":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{record}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/records/r1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": record})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/objects/people/records/r1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":         map[string]any{"record_id": "r1"},
					"created_at": "2025-01-01T00:00:00Z",
					"web_url":    "https://app.attio.com/r/1-updated",
				},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/v2/objects/people/records/r1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":         map[string]any{"record_id": "r1"},
					"created_at": "2025-01-01T00:00:00Z",
					"web_url":    "https://app.attio.com/r/1-replaced",
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/objects/people/records/r1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/records/r1/attributes/email_addresses/values":
			if r.URL.Query().Get("show_historic") != "true" {
				t.Fatalf("expected show_historic query param")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{"active_from": "2025-01-01T00:00:00Z", "value": "ada@example.com"}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/records/r1/entries":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{entry}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "--dry-run", "objects", "create", "--data", `{"api_slug":"people"}`})
	if err != nil {
		t.Fatalf("objects create dry-run failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"dry_run": true`) || strings.Contains(stdout, `"api_slug":"people"`) {
		t.Fatalf("unexpected objects dry-run output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"objects", "get", "people"})
	if err != nil {
		t.Fatalf("objects get failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "SINGULAR") || !strings.Contains(stdout, "people") {
		t.Fatalf("unexpected objects get output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "objects", "update", "people", "--data", `{"singular_noun":"Person Updated"}`})
	if err != nil {
		t.Fatalf("objects update failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "Person Updated") {
		t.Fatalf("unexpected objects update output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "records", "create", "people", "--data", `{"values":{"name":[{"full_name":"Ada"}]}}`})
	if err != nil {
		t.Fatalf("records create failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"record_id": "r1"`) {
		t.Fatalf("unexpected records create output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"records", "assert", "people", "--matching-attribute", "email_addresses", "--data", `{"values":{"name":[{"full_name":"Ada"}]}}`})
	if err != nil {
		t.Fatalf("records assert failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "Ada Lovelace") {
		t.Fatalf("unexpected records assert output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "records", "query", "people", "--all", "--limit", "1", "--filter", `{"name":"Ada"}`, "--sorts", `[{"attribute":"created_at","direction":"asc"}]`})
	if err != nil {
		t.Fatalf("records query --all failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"record_id": "r1"`) {
		t.Fatalf("unexpected records query output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"records", "search", "Ada", "--objects", "people", "--request-as", `{"type":"workspace"}`})
	if err != nil {
		t.Fatalf("records search failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "NAME") || !strings.Contains(stdout, "Ada Lovelace") {
		t.Fatalf("unexpected records search output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "records", "get", "people", "r1"})
	if err != nil {
		t.Fatalf("records get failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"record_id": "r1"`) {
		t.Fatalf("unexpected records get output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "records", "update", "people", "r1", "--data", `{"values":{"name":[{"full_name":"Ada Updated"}]}}`})
	if err != nil {
		t.Fatalf("records update failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "1-updated") {
		t.Fatalf("unexpected records update output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "records", "replace", "people", "r1", "--data", `{"values":{"name":[{"full_name":"Ada Replaced"}]}}`})
	if err != nil {
		t.Fatalf("records replace failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "1-replaced") {
		t.Fatalf("unexpected records replace output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"records", "values", "list", "people", "r1", "email_addresses", "--show-historic", "--limit", "1", "--offset", "2"})
	if err != nil {
		t.Fatalf("records values list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "VALUE") || !strings.Contains(stdout, "ada@example.com") {
		t.Fatalf("unexpected records values output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "records", "entries", "list", "people", "r1", "--limit", "1", "--offset", "3"})
	if err != nil {
		t.Fatalf("records entries list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"entry_id": "e1"`) {
		t.Fatalf("unexpected records entries output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "records", "delete", "people", "r1"})
	if err != nil {
		t.Fatalf("records delete failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"deleted": true`) {
		t.Fatalf("unexpected records delete output: %s", stdout)
	}
}

func TestExecuteNotesTasksCommentsSelfVersionAndCompletionCommands(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/notes":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":            map[string]any{"note_id": "n1"},
					"title":         "Kickoff",
					"created_at":    "2025-01-01T00:00:00Z",
					"parent_record": map[string]any{"record_id": "rec-1"},
				}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/notes/n1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"note_id": "n1"}, "title": "Kickoff"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/notes/n1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/t1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":           map[string]any{"task_id": "t1"},
					"content":      "Follow up",
					"is_completed": false,
					"deadline_at":  "2025-01-01T00:00:00Z",
					"assignees":    []any{map[string]any{"workspace_member_email_address": "owner@example.com"}},
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/tasks/t1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/v2/comments/c1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":         map[string]any{"comment_id": "c1"},
					"thread_id":  map[string]any{"thread_id": "th1"},
					"created_at": "2025-01-01T00:00:00Z",
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/comments/c1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/v2/self":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"active":         true,
				"workspace_name": "Failup",
				"workspace_slug": "failup",
				"workspace_id":   "ws_1",
				"scope":          "workspace:read",
				"client_id":      "cli_1",
				"iat":            1700000000,
				"exp":            1800000000,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"notes", "list", "--parent-object", "people", "--parent-record", "rec-1", "--limit", "1"})
	if err != nil {
		t.Fatalf("notes list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "TITLE") || !strings.Contains(stdout, "Kickoff") {
		t.Fatalf("unexpected notes list output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "notes", "get", "n1"})
	if err != nil {
		t.Fatalf("notes get failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"note_id": "n1"`) {
		t.Fatalf("unexpected notes get output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "notes", "delete", "n1"})
	if err != nil {
		t.Fatalf("notes delete failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"deleted": true`) {
		t.Fatalf("unexpected notes delete output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"tasks", "get", "t1"})
	if err != nil {
		t.Fatalf("tasks get failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "ASSIGNEE") || !strings.Contains(stdout, "owner@example.com") {
		t.Fatalf("unexpected tasks get output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "tasks", "delete", "t1"})
	if err != nil {
		t.Fatalf("tasks delete failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"deleted": true`) {
		t.Fatalf("unexpected tasks delete output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"comments", "get", "c1"})
	if err != nil {
		t.Fatalf("comments get failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "THREAD_ID") || !strings.Contains(stdout, "c1") {
		t.Fatalf("unexpected comments get output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "comments", "delete", "c1"})
	if err != nil {
		t.Fatalf("comments delete failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"deleted": true`) {
		t.Fatalf("unexpected comments delete output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"self"})
	if err != nil {
		t.Fatalf("self failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "workspace_name") || !strings.Contains(stdout, "scope") {
		t.Fatalf("unexpected self table output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"version"})
	if err != nil {
		t.Fatalf("version failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "version\t") || !strings.Contains(stdout, "commit\t") || !strings.Contains(stdout, "date\t") {
		t.Fatalf("unexpected version output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"__complete", "--cword", "1", "attio", "re"})
	if err != nil {
		t.Fatalf("__complete failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "records") {
		t.Fatalf("unexpected __complete output: %s", stdout)
	}
}

func TestExecuteAuthCommandsAndReadKeyFromStdin(t *testing.T) {
	setupCLIEnv(t)

	stdout, stderr, err := captureExecute(t, []string{"--json", "--dry-run", "auth", "login", "--api-key", "secret123"})
	if err != nil {
		t.Fatalf("auth login dry-run failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"dry_run": true`) || strings.Contains(stdout, "secret123") {
		t.Fatalf("unexpected auth dry-run output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"auth", "login", "--api-key", "real-secret-key"})
	if err != nil {
		t.Fatalf("auth login failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "Stored API key") {
		t.Fatalf("unexpected auth login output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"auth", "status"})
	if err != nil {
		t.Fatalf("auth status failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "resolved") || !strings.Contains(stdout, "true") || !strings.Contains(stdout, "resolved_source") || !strings.Contains(stdout, "keyring") {
		t.Fatalf("unexpected auth status output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "auth", "logout"})
	if err != nil {
		t.Fatalf("auth logout failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"removed": true`) {
		t.Fatalf("unexpected auth logout output: %s", stdout)
	}

	origStdin := os.Stdin
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("create stdin pipe: %v", pipeErr)
	}
	_, _ = w.WriteString("  piped-key  \n")
	_ = w.Close()
	os.Stdin = r
	if got := readKeyFromStdin(); got != "piped-key" {
		t.Fatalf("expected piped-key, got %q", got)
	}
	_ = r.Close()

	f, fileErr := os.CreateTemp(t.TempDir(), "stdin-*")
	if fileErr != nil {
		os.Stdin = origStdin
		t.Fatalf("create temp file: %v", fileErr)
	}
	_ = f.Close()
	os.Stdin = f
	if got := readKeyFromStdin(); got != "" {
		os.Stdin = origStdin
		t.Fatalf("expected empty stdin key after read failure, got %q", got)
	}
	os.Stdin = origStdin
}
