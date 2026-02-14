package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func decodeDataEnvelope(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data envelope object, got %#v", body["data"])
	}
	return data
}

func TestExecuteNotesCreateNamedFlags(t *testing.T) {
	setupCLIEnv(t)

	var gotData map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/notes" {
			http.NotFound(w, r)
			return
		}
		gotData = decodeDataEnvelope(t, r)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"note_id": "n1"}, "title": gotData["title"]}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{
		"--json", "notes", "create",
		"--parent-object", "people",
		"--parent-record", "record-1",
		"--title", "Call Summary",
		"--content", "Intro call complete",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if gotData["parent_object"] != "people" {
		t.Fatalf("expected parent_object=people, got %#v", gotData["parent_object"])
	}
	if gotData["parent_record_id"] != "record-1" {
		t.Fatalf("expected parent_record_id=record-1, got %#v", gotData["parent_record_id"])
	}
	if gotData["format"] != "plaintext" {
		t.Fatalf("expected default format plaintext, got %#v", gotData["format"])
	}
	if !strings.Contains(stdout, `"note_id": "n1"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestExecuteTasksCreateNamedFlags(t *testing.T) {
	setupCLIEnv(t)

	var gotData map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/tasks" {
			http.NotFound(w, r)
			return
		}
		gotData = decodeDataEnvelope(t, r)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"task_id": "t1"}, "content": gotData["content"]}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{
		"--json", "tasks", "create",
		"--content", "Follow up",
		"--deadline", "2025-01-01T00:00:00Z",
		"--assignees", "alice@attio.com,member-2",
		"--linked-records", `[{"target_object":"people","target_record_id":"rec-1"}]`,
		"--is-completed", "true",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if gotData["content"] != "Follow up" {
		t.Fatalf("expected content set, got %#v", gotData["content"])
	}
	if gotData["format"] != "plaintext" {
		t.Fatalf("expected format plaintext, got %#v", gotData["format"])
	}
	if gotData["is_completed"] != true {
		t.Fatalf("expected is_completed true, got %#v", gotData["is_completed"])
	}
	assignees, ok := gotData["assignees"].([]any)
	if !ok || len(assignees) != 2 {
		t.Fatalf("expected 2 assignees, got %#v", gotData["assignees"])
	}
	if !strings.Contains(stdout, `"task_id": "t1"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestExecuteTasksCreateDefaultsRequiredFields(t *testing.T) {
	setupCLIEnv(t)

	var gotData map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/tasks" {
			http.NotFound(w, r)
			return
		}
		gotData = decodeDataEnvelope(t, r)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"task_id": "t2"}, "content": gotData["content"]}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	_, stderr, err := captureExecute(t, []string{
		"--json", "tasks", "create",
		"--content", "Minimal task",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if gotData["deadline_at"] != nil {
		t.Fatalf("expected deadline_at default nil, got %#v", gotData["deadline_at"])
	}
	if gotData["is_completed"] != false {
		t.Fatalf("expected is_completed default false, got %#v", gotData["is_completed"])
	}
	if linked, ok := gotData["linked_records"].([]any); !ok || len(linked) != 0 {
		t.Fatalf("expected empty linked_records default, got %#v", gotData["linked_records"])
	}
	if assignees, ok := gotData["assignees"].([]any); !ok || len(assignees) != 0 {
		t.Fatalf("expected empty assignees default, got %#v", gotData["assignees"])
	}
}

func TestExecuteTasksUpdateNamedFlags(t *testing.T) {
	setupCLIEnv(t)

	var gotData map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v2/tasks/task-1" {
			http.NotFound(w, r)
			return
		}
		gotData = decodeDataEnvelope(t, r)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"task_id": "task-1"}}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	_, stderr, err := captureExecute(t, []string{
		"--json", "tasks", "update", "task-1",
		"--deadline", "2025-01-01T00:00:00Z",
		"--assignees", "member-1",
		"--is-completed", "false",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if gotData["is_completed"] != false {
		t.Fatalf("expected is_completed false, got %#v", gotData["is_completed"])
	}
	if _, exists := gotData["content"]; exists {
		t.Fatalf("did not expect content in task update payload: %#v", gotData)
	}
}

func TestExecuteCommentsCreateRecordModeNamedFlags(t *testing.T) {
	setupCLIEnv(t)

	var gotData map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/comments" {
			http.NotFound(w, r)
			return
		}
		gotData = decodeDataEnvelope(t, r)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"comment_id": "c1"}}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	_, stderr, err := captureExecute(t, []string{
		"--json", "comments", "create",
		"--author", "member-1",
		"--body", "Looks good",
		"--record-object", "people",
		"--record-id", "record-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	author, ok := gotData["author"].(map[string]any)
	if !ok {
		t.Fatalf("expected author object, got %#v", gotData["author"])
	}
	if author["type"] != "workspace-member" || author["id"] != "member-1" {
		t.Fatalf("unexpected author payload: %#v", author)
	}
	record, ok := gotData["record"].(map[string]any)
	if !ok {
		t.Fatalf("expected record payload, got %#v", gotData["record"])
	}
	if record["object"] != "people" || record["record_id"] != "record-1" {
		t.Fatalf("unexpected record payload: %#v", record)
	}
}

func TestExecuteCommentsCreateRequiresTargetMode(t *testing.T) {
	setupCLIEnv(t)
	t.Setenv("ATTIO_API_KEY", "env-key")

	_, stderr, err := captureExecute(t, []string{"comments", "create", "--author", "member-1", "--body", "hello"})
	if err == nil {
		t.Fatalf("expected usage error")
	}
	if ExitCode(err) != ExitCodeUsage {
		t.Fatalf("expected usage exit code %d, got %d", ExitCodeUsage, ExitCode(err))
	}
	if !strings.Contains(stderr, "one of") {
		t.Fatalf("expected target-mode guidance, got %q", stderr)
	}
}

func TestExecuteWebhooksCreateNamedFlags(t *testing.T) {
	setupCLIEnv(t)

	var gotData map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/webhooks" {
			http.NotFound(w, r)
			return
		}
		gotData = decodeDataEnvelope(t, r)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"webhook_id": "w1"}, "target_url": gotData["target_url"]}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	_, stderr, err := captureExecute(t, []string{
		"--json", "webhooks", "create",
		"--target-url", "https://example.com/webhook",
		"--subscriptions", `[{"event_type":"note.created","filter":null}]`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if gotData["target_url"] != "https://example.com/webhook" {
		t.Fatalf("expected target_url set, got %#v", gotData["target_url"])
	}
	subs, ok := gotData["subscriptions"].([]any)
	if !ok || len(subs) != 1 {
		t.Fatalf("expected one subscription, got %#v", gotData["subscriptions"])
	}
}

func TestExecuteRecordsQueryTableShowsNameAndEmail(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/objects/people/records/query" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
			"id":         map[string]any{"record_id": "r1"},
			"created_at": "2025-01-01T00:00:00Z",
			"web_url":    "https://app.attio.com/r/1",
			"values": map[string]any{
				"name":            []any{map[string]any{"full_name": "Ada Lovelace"}},
				"email_addresses": []any{map[string]any{"email_address": "ada@example.com"}},
			},
		}}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"records", "query", "people", "--limit", "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "NAME") || !strings.Contains(stdout, "EMAIL") {
		t.Fatalf("expected NAME/EMAIL table columns, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Ada Lovelace") || !strings.Contains(stdout, "ada@example.com") {
		t.Fatalf("expected extracted name/email values, got: %s", stdout)
	}
}

func TestExecuteTasksListTableShowsStatusAndAssignee(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/tasks" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
			"id":           map[string]any{"task_id": "t1"},
			"content":      "Follow up",
			"is_completed": true,
			"deadline_at":  "2025-01-01T00:00:00Z",
			"assignees":    []any{map[string]any{"workspace_member_email_address": "alice@attio.com"}},
		}}})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"tasks", "list", "--limit", "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "STATUS") || !strings.Contains(stdout, "ASSIGNEE") {
		t.Fatalf("expected STATUS/ASSIGNEE columns, got: %s", stdout)
	}
	if !strings.Contains(stdout, "completed") || !strings.Contains(stdout, "alice@attio.com") {
		t.Fatalf("expected completed status and assignee value, got: %s", stdout)
	}
}

func TestExecuteMeetingsListTableShowsParticipants(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/meetings" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{
				"id":           map[string]any{"meeting_id": "m1"},
				"title":        "Pipeline Review",
				"start_at":     "2025-01-01T10:00:00Z",
				"end_at":       "2025-01-01T10:30:00Z",
				"participants": []any{map[string]any{"name": "Ada"}, map[string]any{"name": "Lin"}},
			}},
			"pagination": map[string]any{"next_cursor": ""},
		})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"meetings", "list", "--limit", "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "PARTICIPANTS") {
		t.Fatalf("expected PARTICIPANTS column, got: %s", stdout)
	}
	if !strings.Contains(stdout, "2") {
		t.Fatalf("expected participant count in output, got: %s", stdout)
	}
}
