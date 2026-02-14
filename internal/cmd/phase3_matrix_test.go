package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExecuteListsCommands(t *testing.T) {
	setupCLIEnv(t)

	listObject := map[string]any{
		"id":            map[string]any{"list_id": "l1"},
		"api_slug":      "prospects",
		"name":          "Prospects",
		"parent_object": map[string]any{"api_slug": "people"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/lists":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{listObject}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/lists":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": listObject})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/lists/prospects":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": listObject})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/lists/prospects":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"list_id": "l1"}, "api_slug": "prospects", "name": "Prospects Updated", "parent_object": map[string]any{"api_slug": "people"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "lists", "list"})
	if err != nil {
		t.Fatalf("lists list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"api_slug": "prospects"`) {
		t.Fatalf("unexpected lists list output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "lists", "create", "--data", `{"api_slug":"prospects"}`})
	if err != nil {
		t.Fatalf("lists create failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"list_id": "l1"`) {
		t.Fatalf("unexpected lists create output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"lists", "get", "prospects"})
	if err != nil {
		t.Fatalf("lists get failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "ID") || !strings.Contains(stdout, "prospects") {
		t.Fatalf("unexpected lists get table output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "lists", "update", "prospects", "--data", `{"name":"Prospects Updated"}`})
	if err != nil {
		t.Fatalf("lists update failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "Prospects Updated") {
		t.Fatalf("unexpected lists update output: %s", stdout)
	}
}

func TestExecuteEntriesCommands(t *testing.T) {
	setupCLIEnv(t)

	entryObject := map[string]any{
		"id":         map[string]any{"entry_id": "e1"},
		"created_at": "2025-01-01T00:00:00Z",
		"web_url":    "https://app.attio.com/e/1",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2/lists/prospects/entries":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": entryObject})
		case r.Method == http.MethodPut && r.URL.Path == "/v2/lists/prospects/entries":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": entryObject})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/lists/prospects/entries/query":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			offset, _ := body["offset"].(float64)
			if offset == 0 {
				_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{entryObject}})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/lists/prospects/entries/e1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": entryObject})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/lists/prospects/entries/e1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"entry_id": "e1"}, "created_at": "2025-01-01T00:00:00Z", "web_url": "https://app.attio.com/e/1-updated"}})
		case r.Method == http.MethodPut && r.URL.Path == "/v2/lists/prospects/entries/e1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"entry_id": "e1"}, "created_at": "2025-01-01T00:00:00Z", "web_url": "https://app.attio.com/e/1-replaced"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/lists/prospects/entries/e1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/v2/lists/prospects/entries/e1/attributes/stage/values":
			if r.URL.Query().Get("show_historic") != "true" {
				t.Fatalf("expected show_historic query param")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"active_from": "2025-01-01T00:00:00Z", "value": "Open"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "entries", "create", "prospects", "--data", `{"parent_record_id":"r1"}`})
	if err != nil {
		t.Fatalf("entries create failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"entry_id": "e1"`) {
		t.Fatalf("unexpected entries create output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "entries", "assert", "prospects", "--data", `{"parent_record_id":"r1"}`})
	if err != nil {
		t.Fatalf("entries assert failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"entry_id": "e1"`) {
		t.Fatalf("unexpected entries assert output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"entries", "query", "prospects", "--all", "--limit", "1"})
	if err != nil {
		t.Fatalf("entries query failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "WEB_URL") || !strings.Contains(stdout, "e1") {
		t.Fatalf("unexpected entries query table output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "entries", "get", "prospects", "e1"})
	if err != nil {
		t.Fatalf("entries get failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"entry_id": "e1"`) {
		t.Fatalf("unexpected entries get output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "entries", "update", "prospects", "e1", "--data", `{"stage":"updated"}`})
	if err != nil {
		t.Fatalf("entries update failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "1-updated") {
		t.Fatalf("unexpected entries update output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "entries", "replace", "prospects", "e1", "--data", `{"stage":"replaced"}`})
	if err != nil {
		t.Fatalf("entries replace failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "1-replaced") {
		t.Fatalf("unexpected entries replace output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "entries", "values", "list", "prospects", "e1", "stage", "--show-historic", "--limit", "1", "--offset", "2"})
	if err != nil {
		t.Fatalf("entries values list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"value": "Open"`) {
		t.Fatalf("unexpected entries values output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "entries", "delete", "prospects", "e1"})
	if err != nil {
		t.Fatalf("entries delete failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"deleted": true`) {
		t.Fatalf("unexpected entries delete output: %s", stdout)
	}
}

func TestExecuteThreadsMembersWebhooksCommands(t *testing.T) {
	setupCLIEnv(t)

	threadObject := map[string]any{"id": map[string]any{"thread_id": "th1"}, "is_resolved": false, "created_at": "2025-01-01T00:00:00Z"}
	memberObject := map[string]any{"id": map[string]any{"workspace_member_id": "m1"}, "name": "Member One", "email_address": "m1@example.com", "role": "admin"}
	webhookObject := map[string]any{"id": map[string]any{"webhook_id": "wh1"}, "target_url": "https://example.com/hook", "status": "active", "subscriptions": []any{map[string]any{"event_type": "task.created", "filter": nil}}, "created_at": "2025-01-01T00:00:00Z"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/threads":
			if r.URL.Query().Get("record_id") != "rec-1" || r.URL.Query().Get("object") != "people" {
				t.Fatalf("unexpected threads list query: %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{threadObject}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/threads/th1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": threadObject})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/workspace_members/m1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": memberObject})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/webhooks":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{webhookObject}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/webhooks/wh1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": webhookObject})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/webhooks/wh1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"webhook_id": "wh1"}, "target_url": "https://example.com/updated", "status": "active", "subscriptions": []any{map[string]any{"event_type": "task.updated", "filter": nil}}, "created_at": "2025-01-01T00:00:00Z"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/webhooks/wh1":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"threads", "list", "--record-id", "rec-1", "--object", "people"})
	if err != nil {
		t.Fatalf("threads list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "IS_RESOLVED") || !strings.Contains(stdout, "th1") {
		t.Fatalf("unexpected threads list output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "threads", "get", "th1"})
	if err != nil {
		t.Fatalf("threads get failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"thread_id": "th1"`) {
		t.Fatalf("unexpected threads get output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"members", "get", "m1"})
	if err != nil {
		t.Fatalf("members get failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "Member One") || !strings.Contains(stdout, "EMAIL") {
		t.Fatalf("unexpected members get output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"webhooks", "list"})
	if err != nil {
		t.Fatalf("webhooks list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "SUBSCRIPTIONS") || !strings.Contains(stdout, "wh1") {
		t.Fatalf("unexpected webhooks list output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "webhooks", "get", "wh1"})
	if err != nil {
		t.Fatalf("webhooks get failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"webhook_id": "wh1"`) {
		t.Fatalf("unexpected webhooks get output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "webhooks", "update", "wh1", "--target-url", "https://example.com/updated", "--subscriptions", `[{"event_type":"task.updated","filter":null}]`})
	if err != nil {
		t.Fatalf("webhooks update failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "task.updated") {
		t.Fatalf("unexpected webhooks update output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "webhooks", "delete", "wh1"})
	if err != nil {
		t.Fatalf("webhooks delete failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"deleted": true`) {
		t.Fatalf("unexpected webhooks delete output: %s", stdout)
	}
}

func TestExecuteWebhooksUpdateRequiresPayload(t *testing.T) {
	setupCLIEnv(t)
	t.Setenv("ATTIO_API_KEY", "env-key")

	_, stderr, err := captureExecute(t, []string{"webhooks", "update", "wh1"})
	if err == nil {
		t.Fatalf("expected usage error")
	}
	if ExitCode(err) != ExitCodeUsage {
		t.Fatalf("expected usage exit code %d, got %d", ExitCodeUsage, ExitCode(err))
	}
	if !strings.Contains(stderr, "requires at least one update field") {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestExecuteMeetingsCommands(t *testing.T) {
	setupCLIEnv(t)

	next := "next"
	meetingObject := map[string]any{"id": map[string]any{"meeting_id": "m1"}, "title": "Pipeline Review", "start_at": "2025-01-01T10:00:00Z", "end_at": "2025-01-01T10:30:00Z", "participants": []any{map[string]any{"name": "Ada"}}}
	recordingObject := map[string]any{"id": map[string]any{"call_recording_id": "r1"}, "created_at": "2025-01-01T10:00:00Z", "url": "https://recording.example/r1"}
	segment := map[string]any{"start_at": "0", "end_at": "1", "speaker_name": "Ada", "text": "Hello"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings/m1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": meetingObject})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/meetings":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": meetingObject})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings/m1/call_recordings":
			cursor := r.URL.Query().Get("cursor")
			if cursor == "" {
				_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{recordingObject}, "pagination": map[string]any{"next_cursor": next}})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}, "pagination": map[string]any{"next_cursor": ""}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings/m1/call_recordings/r1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": recordingObject})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/meetings/m1/call_recordings":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": recordingObject})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/meetings/m1/call_recordings/r1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings/m1/call_recordings/r1/transcript":
			cursor := r.URL.Query().Get("cursor")
			if cursor == "" {
				_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{segment}, "pagination": map[string]any{"next_cursor": next}})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}, "pagination": map[string]any{"next_cursor": ""}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"meetings", "get", "m1"})
	if err != nil {
		t.Fatalf("meetings get failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "PARTICIPANTS") || !strings.Contains(stdout, "Pipeline Review") {
		t.Fatalf("unexpected meetings get output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "meetings", "create", "--data", `{"external_ref":"abc"}`})
	if err != nil {
		t.Fatalf("meetings create failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"meeting_id": "m1"`) {
		t.Fatalf("unexpected meetings create output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"meetings", "recordings", "list", "m1", "--limit", "1"})
	if err != nil {
		t.Fatalf("meetings recordings list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "CREATED_AT") || !strings.Contains(stdout, "r1") {
		t.Fatalf("unexpected meetings recordings list output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "meetings", "recordings", "list", "m1", "--all", "--limit", "1"})
	if err != nil {
		t.Fatalf("meetings recordings list --all failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"call_recording_id": "r1"`) {
		t.Fatalf("unexpected meetings recordings list --all output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "meetings", "recordings", "get", "m1", "r1"})
	if err != nil {
		t.Fatalf("meetings recordings get failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "recording.example") {
		t.Fatalf("unexpected meetings recordings get output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "meetings", "recordings", "create", "m1", "--data", `{"url":"https://recording.example/r1"}`})
	if err != nil {
		t.Fatalf("meetings recordings create failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"call_recording_id": "r1"`) {
		t.Fatalf("unexpected meetings recordings create output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "meetings", "recordings", "delete", "m1", "r1"})
	if err != nil {
		t.Fatalf("meetings recordings delete failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"deleted": true`) {
		t.Fatalf("unexpected meetings recordings delete output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"meetings", "transcript", "m1", "r1"})
	if err != nil {
		t.Fatalf("meetings transcript failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "SPEAKER") || !strings.Contains(stdout, "Hello") {
		t.Fatalf("unexpected meetings transcript table output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "meetings", "transcript", "m1", "r1", "--all"})
	if err != nil {
		t.Fatalf("meetings transcript --all failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"speaker_name": "Ada"`) {
		t.Fatalf("unexpected meetings transcript --all output: %s", stdout)
	}
}
