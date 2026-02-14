package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExecuteRuntimeErrorMatrix(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status_code":400,"type":"validation_error","code":"invalid_request","message":"bad request"}`))
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	cases := []struct {
		name string
		args []string
	}{
		{name: "objects-list", args: []string{"objects", "list"}},
		{name: "objects-get", args: []string{"objects", "get", "people"}},
		{name: "objects-create", args: []string{"objects", "create", "--data", `{"api_slug":"people"}`}},
		{name: "objects-update", args: []string{"objects", "update", "people", "--data", `{"singular_noun":"Person"}`}},
		{name: "lists-list", args: []string{"lists", "list"}},
		{name: "lists-get", args: []string{"lists", "get", "prospects"}},
		{name: "lists-create", args: []string{"lists", "create", "--data", `{"api_slug":"prospects"}`}},
		{name: "lists-update", args: []string{"lists", "update", "prospects", "--data", `{"name":"Prospects"}`}},
		{name: "records-create", args: []string{"records", "create", "people", "--data", `{"values":{}}`}},
		{name: "records-assert", args: []string{"records", "assert", "people", "--matching-attribute", "email_addresses", "--data", `{"values":{}}`}},
		{name: "records-query", args: []string{"records", "query", "people", "--limit", "1"}},
		{name: "records-search", args: []string{"records", "search", "ada", "--objects", "people"}},
		{name: "records-get", args: []string{"records", "get", "people", "rec_1"}},
		{name: "records-update", args: []string{"records", "update", "people", "rec_1", "--data", `{"values":{}}`}},
		{name: "records-replace", args: []string{"records", "replace", "people", "rec_1", "--data", `{"values":{}}`}},
		{name: "records-delete", args: []string{"records", "delete", "people", "rec_1"}},
		{name: "records-values-list", args: []string{"records", "values", "list", "people", "rec_1", "email_addresses", "--limit", "1"}},
		{name: "records-entries-list", args: []string{"records", "entries", "list", "people", "rec_1", "--limit", "1"}},
		{name: "entries-create", args: []string{"entries", "create", "prospects", "--data", `{"values":{}}`}},
		{name: "entries-assert", args: []string{"entries", "assert", "prospects", "--data", `{"values":{}}`}},
		{name: "entries-query", args: []string{"entries", "query", "prospects", "--limit", "1"}},
		{name: "entries-get", args: []string{"entries", "get", "prospects", "entry_1"}},
		{name: "entries-update", args: []string{"entries", "update", "prospects", "entry_1", "--data", `{"values":{}}`}},
		{name: "entries-replace", args: []string{"entries", "replace", "prospects", "entry_1", "--data", `{"values":{}}`}},
		{name: "entries-delete", args: []string{"entries", "delete", "prospects", "entry_1"}},
		{name: "entries-values-list", args: []string{"entries", "values", "list", "prospects", "entry_1", "status", "--limit", "1"}},
		{name: "notes-list", args: []string{"notes", "list", "--parent-object", "people", "--parent-record", "rec_1", "--limit", "1"}},
		{name: "notes-create", args: []string{"notes", "create", "--parent-object", "people", "--parent-record", "rec_1", "--title", "Note", "--content", "Body"}},
		{name: "notes-get", args: []string{"notes", "get", "note_1"}},
		{name: "notes-delete", args: []string{"notes", "delete", "note_1"}},
		{name: "tasks-list", args: []string{"tasks", "list", "--limit", "1"}},
		{name: "tasks-create", args: []string{"tasks", "create", "--content", "Follow up"}},
		{name: "tasks-get", args: []string{"tasks", "get", "task_1"}},
		{name: "tasks-update", args: []string{"tasks", "update", "task_1", "--is-completed", "true"}},
		{name: "tasks-delete", args: []string{"tasks", "delete", "task_1"}},
		{name: "comments-create", args: []string{"comments", "create", "--author", "member_1", "--content", "Hello", "--thread", "thread_1"}},
		{name: "comments-get", args: []string{"comments", "get", "comment_1"}},
		{name: "comments-delete", args: []string{"comments", "delete", "comment_1"}},
		{name: "threads-list", args: []string{"threads", "list", "--object", "people", "--record-id", "rec_1", "--limit", "1"}},
		{name: "threads-get", args: []string{"threads", "get", "thread_1"}},
		{name: "meetings-list", args: []string{"meetings", "list", "--limit", "1"}},
		{name: "meetings-get", args: []string{"meetings", "get", "meeting_1"}},
		{name: "meetings-create", args: []string{"meetings", "create", "--data", `{"title":"Sync"}`}},
		{name: "recordings-list", args: []string{"meetings", "recordings", "list", "meeting_1", "--limit", "1"}},
		{name: "recordings-get", args: []string{"meetings", "recordings", "get", "meeting_1", "recording_1"}},
		{name: "recordings-create", args: []string{"meetings", "recordings", "create", "meeting_1", "--data", `{"url":"https://example.com"}`}},
		{name: "recordings-delete", args: []string{"meetings", "recordings", "delete", "meeting_1", "recording_1"}},
		{name: "transcript", args: []string{"meetings", "transcript", "meeting_1", "recording_1"}},
		{name: "webhooks-list", args: []string{"webhooks", "list", "--limit", "1"}},
		{name: "webhooks-get", args: []string{"webhooks", "get", "webhook_1"}},
		{name: "webhooks-create", args: []string{"webhooks", "create", "--target-url", "https://example.com", "--subscriptions", `[{"event_type":"note.created","filter":null}]`}},
		{name: "webhooks-update", args: []string{"webhooks", "update", "webhook_1", "--target-url", "https://example.com"}},
		{name: "webhooks-delete", args: []string{"webhooks", "delete", "webhook_1"}},
		{name: "members-list", args: []string{"members", "list"}},
		{name: "members-get", args: []string{"members", "get", "member_1"}},
		{name: "attributes-list", args: []string{"attributes", "list", "objects", "people", "--limit", "1"}},
		{name: "attributes-get", args: []string{"attributes", "get", "objects", "people", "stage"}},
		{name: "attributes-create", args: []string{"attributes", "create", "objects", "people", "--data", `{"title":"Stage"}`}},
		{name: "attributes-update", args: []string{"attributes", "update", "objects", "people", "stage", "--data", `{"title":"Stage 2"}`}},
		{name: "options-list", args: []string{"attributes", "options", "list", "objects", "people", "stage"}},
		{name: "options-create", args: []string{"attributes", "options", "create", "objects", "people", "stage", "--data", `{"title":"Hot"}`}},
		{name: "options-update", args: []string{"attributes", "options", "update", "objects", "people", "stage", "hot", "--data", `{"title":"Warm"}`}},
		{name: "statuses-list", args: []string{"attributes", "statuses", "list", "objects", "people", "stage"}},
		{name: "statuses-create", args: []string{"attributes", "statuses", "create", "objects", "people", "stage", "--data", `{"title":"Open"}`}},
		{name: "statuses-update", args: []string{"attributes", "statuses", "update", "objects", "people", "stage", "open", "--data", `{"title":"Closed"}`}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, stderr, err := captureExecute(t, tc.args)
			if err == nil {
				t.Fatalf("expected runtime error")
			}
			if ExitCode(err) != ExitCodeGeneric {
				t.Fatalf("expected runtime exit code %d, got %d (stderr=%s)", ExitCodeGeneric, ExitCode(err), stderr)
			}
			if !strings.Contains(stderr, "Attio API error") {
				t.Fatalf("expected API error in stderr, got: %s", stderr)
			}
		})
	}
}
