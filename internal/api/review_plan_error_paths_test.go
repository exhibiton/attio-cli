package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIMethodsReturnAttioErrorOnServerFailure(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status_code":400,"type":"validation_error","code":"invalid_request","message":"bad request"}`))
	}))
	defer srv.Close()

	client := NewClient("test-key", srv.URL)
	ctx := context.Background()

	isCompleted := true
	cases := []struct {
		name string
		call func() error
	}{
		{name: "get-self", call: func() error { _, err := client.GetSelf(ctx); return err }},
		{name: "list-objects", call: func() error { _, err := client.ListObjects(ctx); return err }},
		{name: "create-object", call: func() error { _, err := client.CreateObject(ctx, map[string]any{"api_slug": "people"}); return err }},
		{name: "get-object", call: func() error { _, err := client.GetObject(ctx, "people"); return err }},
		{name: "update-object", call: func() error {
			_, err := client.UpdateObject(ctx, "people", map[string]any{"singular_noun": "Person"})
			return err
		}},
		{name: "list-lists", call: func() error { _, err := client.ListLists(ctx); return err }},
		{name: "create-list", call: func() error { _, err := client.CreateList(ctx, map[string]any{"api_slug": "prospects"}); return err }},
		{name: "get-list", call: func() error { _, err := client.GetList(ctx, "prospects"); return err }},
		{name: "update-list", call: func() error {
			_, err := client.UpdateList(ctx, "prospects", map[string]any{"name": "Prospects"})
			return err
		}},
		{name: "create-record", call: func() error {
			_, err := client.CreateRecord(ctx, "people", map[string]any{"values": map[string]any{}})
			return err
		}},
		{name: "assert-record", call: func() error {
			_, err := client.AssertRecord(ctx, "people", "email_addresses", map[string]any{"values": map[string]any{}})
			return err
		}},
		{name: "query-records", call: func() error { _, err := client.QueryRecords(ctx, "people", nil, nil, 1, 0); return err }},
		{name: "search-records", call: func() error {
			_, err := client.SearchRecords(ctx, "ada", 1, []string{"people"}, map[string]any{"type": "workspace"})
			return err
		}},
		{name: "get-record", call: func() error { _, err := client.GetRecord(ctx, "people", "rec_1"); return err }},
		{name: "update-record", call: func() error {
			_, err := client.UpdateRecord(ctx, "people", "rec_1", map[string]any{"values": map[string]any{}})
			return err
		}},
		{name: "replace-record", call: func() error {
			_, err := client.ReplaceRecord(ctx, "people", "rec_1", map[string]any{"values": map[string]any{}})
			return err
		}},
		{name: "delete-record", call: func() error { return client.DeleteRecord(ctx, "people", "rec_1") }},
		{name: "list-record-values", call: func() error {
			_, err := client.ListRecordAttributeValues(ctx, "people", "rec_1", "email_addresses", false, 1, 0)
			return err
		}},
		{name: "list-record-entries", call: func() error { _, err := client.ListRecordEntries(ctx, "people", "rec_1", 1, 0); return err }},
		{name: "create-entry", call: func() error {
			_, err := client.CreateEntry(ctx, "prospects", map[string]any{"values": map[string]any{}})
			return err
		}},
		{name: "assert-entry", call: func() error {
			_, err := client.AssertEntry(ctx, "prospects", map[string]any{"values": map[string]any{}})
			return err
		}},
		{name: "query-entries", call: func() error { _, err := client.QueryEntries(ctx, "prospects", nil, nil, 1, 0); return err }},
		{name: "get-entry", call: func() error { _, err := client.GetEntry(ctx, "prospects", "entry_1"); return err }},
		{name: "update-entry", call: func() error {
			_, err := client.UpdateEntry(ctx, "prospects", "entry_1", map[string]any{"values": map[string]any{}})
			return err
		}},
		{name: "replace-entry", call: func() error {
			_, err := client.ReplaceEntry(ctx, "prospects", "entry_1", map[string]any{"values": map[string]any{}})
			return err
		}},
		{name: "delete-entry", call: func() error { return client.DeleteEntry(ctx, "prospects", "entry_1") }},
		{name: "list-entry-values", call: func() error {
			_, err := client.ListEntryAttributeValues(ctx, "prospects", "entry_1", "status", false, 1, 0)
			return err
		}},
		{name: "list-notes", call: func() error { _, err := client.ListNotes(ctx, "people", "rec_1", 1, 0); return err }},
		{name: "create-note", call: func() error { _, err := client.CreateNote(ctx, map[string]any{"title": "Title"}); return err }},
		{name: "get-note", call: func() error { _, err := client.GetNote(ctx, "note_1"); return err }},
		{name: "delete-note", call: func() error { return client.DeleteNote(ctx, "note_1") }},
		{name: "list-tasks", call: func() error { _, err := client.ListTasks(ctx, 1, 0, "", "", "", "", &isCompleted); return err }},
		{name: "create-task", call: func() error { _, err := client.CreateTask(ctx, map[string]any{"content": "Follow up"}); return err }},
		{name: "get-task", call: func() error { _, err := client.GetTask(ctx, "task_1"); return err }},
		{name: "update-task", call: func() error {
			_, err := client.UpdateTask(ctx, "task_1", map[string]any{"is_completed": true})
			return err
		}},
		{name: "delete-task", call: func() error { return client.DeleteTask(ctx, "task_1") }},
		{name: "create-comment", call: func() error { _, err := client.CreateComment(ctx, map[string]any{"content": "Hi"}); return err }},
		{name: "get-comment", call: func() error { _, err := client.GetComment(ctx, "comment_1"); return err }},
		{name: "delete-comment", call: func() error { return client.DeleteComment(ctx, "comment_1") }},
		{name: "list-threads", call: func() error { _, err := client.ListThreads(ctx, "people", "rec_1", "", "", 1, 0); return err }},
		{name: "get-thread", call: func() error { _, err := client.GetThread(ctx, "thread_1"); return err }},
		{name: "list-meetings", call: func() error { _, _, err := client.ListMeetings(ctx, 1, "", "", "", "", "", "", "", ""); return err }},
		{name: "find-or-create-meeting", call: func() error { _, err := client.FindOrCreateMeeting(ctx, map[string]any{"title": "Call"}); return err }},
		{name: "get-meeting", call: func() error { _, err := client.GetMeeting(ctx, "meeting_1"); return err }},
		{name: "list-recordings", call: func() error { _, _, err := client.ListCallRecordings(ctx, "meeting_1", 1, ""); return err }},
		{name: "create-recording", call: func() error {
			_, err := client.CreateCallRecording(ctx, "meeting_1", map[string]any{"url": "https://example.com"})
			return err
		}},
		{name: "get-recording", call: func() error { _, err := client.GetCallRecording(ctx, "meeting_1", "recording_1"); return err }},
		{name: "delete-recording", call: func() error { return client.DeleteCallRecording(ctx, "meeting_1", "recording_1") }},
		{name: "get-transcript", call: func() error { _, _, err := client.GetTranscript(ctx, "meeting_1", "recording_1", ""); return err }},
		{name: "list-webhooks", call: func() error { _, err := client.ListWebhooks(ctx, 1, 0); return err }},
		{name: "create-webhook", call: func() error {
			_, err := client.CreateWebhook(ctx, map[string]any{"target_url": "https://example.com"})
			return err
		}},
		{name: "get-webhook", call: func() error { _, err := client.GetWebhook(ctx, "webhook_1"); return err }},
		{name: "update-webhook", call: func() error {
			_, err := client.UpdateWebhook(ctx, "webhook_1", map[string]any{"status": "active"})
			return err
		}},
		{name: "delete-webhook", call: func() error { return client.DeleteWebhook(ctx, "webhook_1") }},
		{name: "list-members", call: func() error { _, err := client.ListMembers(ctx); return err }},
		{name: "get-member", call: func() error { _, err := client.GetMember(ctx, "member_1"); return err }},
		{name: "list-attributes", call: func() error { _, err := client.ListAttributes(ctx, "objects", "people", false, 1, 0); return err }},
		{name: "create-attribute", call: func() error {
			_, err := client.CreateAttribute(ctx, "objects", "people", map[string]any{"title": "Stage"})
			return err
		}},
		{name: "get-attribute", call: func() error { _, err := client.GetAttribute(ctx, "objects", "people", "stage"); return err }},
		{name: "update-attribute", call: func() error {
			_, err := client.UpdateAttribute(ctx, "objects", "people", "stage", map[string]any{"title": "New Stage"})
			return err
		}},
		{name: "list-select-options", call: func() error { _, err := client.ListSelectOptions(ctx, "objects", "people", "stage", false); return err }},
		{name: "create-select-option", call: func() error {
			_, err := client.CreateSelectOption(ctx, "objects", "people", "stage", map[string]any{"title": "Hot"})
			return err
		}},
		{name: "update-select-option", call: func() error {
			_, err := client.UpdateSelectOption(ctx, "objects", "people", "stage", "hot", map[string]any{"title": "Warm"})
			return err
		}},
		{name: "list-statuses", call: func() error { _, err := client.ListStatuses(ctx, "objects", "people", "stage", false); return err }},
		{name: "create-status", call: func() error {
			_, err := client.CreateStatus(ctx, "objects", "people", "stage", map[string]any{"title": "Open"})
			return err
		}},
		{name: "update-status", call: func() error {
			_, err := client.UpdateStatus(ctx, "objects", "people", "stage", "open", map[string]any{"title": "Closed"})
			return err
		}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			if err == nil {
				t.Fatalf("expected error")
			}

			var attioErr *AttioError
			if !errors.As(err, &attioErr) {
				t.Fatalf("expected AttioError, got %T", err)
			}
			if attioErr.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", attioErr.StatusCode)
			}
		})
	}
}

func TestNewClientDefaultsAndAuthHelpers(t *testing.T) {
	t.Parallel()

	client := NewClient("test-key", "")
	if client.baseURL != defaultBaseURL {
		t.Fatalf("expected default base URL %q, got %q", defaultBaseURL, client.baseURL)
	}
	if client.userAgent != "attio-cli/dev" {
		t.Fatalf("unexpected default user-agent: %q", client.userAgent)
	}
	if client.httpClient.Timeout.Seconds() <= 0 {
		t.Fatalf("expected positive default timeout, got %v", client.httpClient.Timeout)
	}

	if !IsAuthError(&AttioError{StatusCode: http.StatusUnauthorized}) {
		t.Fatalf("expected unauthorized AttioError to be auth error")
	}
	if !IsAuthError(&AttioError{StatusCode: http.StatusForbidden}) {
		t.Fatalf("expected forbidden AttioError to be auth error")
	}
	if IsAuthError(errors.New("nope")) {
		t.Fatalf("did not expect generic error to be auth error")
	}
}
