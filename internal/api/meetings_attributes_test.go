package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMeetingsAndAttributesAPI(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings":
			cursor := r.URL.Query().Get("cursor")
			if cursor == "" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data":       []map[string]any{{"id": map[string]any{"meeting_id": "m1"}, "title": "Pipeline Review"}},
					"pagination": map[string]any{"next_cursor": "c1"},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}, "pagination": map[string]any{"next_cursor": nil}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/meetings":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"meeting_id": "m1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings/m1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"meeting_id": "m1"}, "title": "Pipeline Review"}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings/m1/call_recordings":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"call_recording_id": "cr1"}}}, "pagination": map[string]any{"next_cursor": nil}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/meetings/m1/call_recordings":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"call_recording_id": "cr1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings/m1/call_recordings/cr1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"call_recording_id": "cr1"}}})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/meetings/m1/call_recordings/cr1":
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings/m1/call_recordings/cr1/transcript":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"text": "Hello"}}, "pagination": map[string]any{"next_cursor": nil}})

		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/attributes":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"attribute_id": "a1"}, "title": "Stage"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/people/attributes":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"attribute_id": "a1"}, "title": "Stage"}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/attributes/a1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"attribute_id": "a1"}, "title": "Stage"}})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/objects/people/attributes/a1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"attribute_id": "a1"}, "title": "Stage Updated"}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/attributes/a1/options":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"option_id": "o1"}, "title": "Hot"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/people/attributes/a1/options":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"option_id": "o1"}, "title": "Hot"}})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/objects/people/attributes/a1/options/o1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"option_id": "o1"}, "title": "Warm"}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/attributes/a1/statuses":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"status_id": "s1"}, "title": "Open"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/people/attributes/a1/statuses":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"status_id": "s1"}, "title": "Open"}})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/objects/people/attributes/a1/statuses/s1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"status_id": "s1"}, "title": "Closed"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient("test-key", srv.URL)

	meetings, next, err := client.ListMeetings(context.Background(), 10, "", "", "", "", "", "", "", "")
	if err != nil || len(meetings) != 1 || next != "c1" {
		t.Fatalf("list meetings failed: err=%v len=%d next=%q", err, len(meetings), next)
	}
	if _, err := client.FindOrCreateMeeting(context.Background(), map[string]any{"title": "Pipeline Review"}); err != nil {
		t.Fatalf("find or create meeting: %v", err)
	}
	if _, err := client.GetMeeting(context.Background(), "m1"); err != nil {
		t.Fatalf("get meeting: %v", err)
	}
	if _, _, err := client.ListCallRecordings(context.Background(), "m1", 10, ""); err != nil {
		t.Fatalf("list call recordings: %v", err)
	}
	if _, err := client.CreateCallRecording(context.Background(), "m1", map[string]any{"url": "https://example.com"}); err != nil {
		t.Fatalf("create call recording: %v", err)
	}
	if _, err := client.GetCallRecording(context.Background(), "m1", "cr1"); err != nil {
		t.Fatalf("get call recording: %v", err)
	}
	if err := client.DeleteCallRecording(context.Background(), "m1", "cr1"); err != nil {
		t.Fatalf("delete call recording: %v", err)
	}
	if _, _, err := client.GetTranscript(context.Background(), "m1", "cr1", ""); err != nil {
		t.Fatalf("get transcript: %v", err)
	}

	if _, err := client.ListAttributes(context.Background(), "objects", "people", false, 10, 0); err != nil {
		t.Fatalf("list attributes: %v", err)
	}
	if _, err := client.CreateAttribute(context.Background(), "objects", "people", map[string]any{"title": "Stage"}); err != nil {
		t.Fatalf("create attribute: %v", err)
	}
	if _, err := client.GetAttribute(context.Background(), "objects", "people", "a1"); err != nil {
		t.Fatalf("get attribute: %v", err)
	}
	if _, err := client.UpdateAttribute(context.Background(), "objects", "people", "a1", map[string]any{"title": "Stage Updated"}); err != nil {
		t.Fatalf("update attribute: %v", err)
	}
	if _, err := client.ListSelectOptions(context.Background(), "objects", "people", "a1", false); err != nil {
		t.Fatalf("list options: %v", err)
	}
	if _, err := client.CreateSelectOption(context.Background(), "objects", "people", "a1", map[string]any{"title": "Hot"}); err != nil {
		t.Fatalf("create option: %v", err)
	}
	if _, err := client.UpdateSelectOption(context.Background(), "objects", "people", "a1", "o1", map[string]any{"title": "Warm"}); err != nil {
		t.Fatalf("update option: %v", err)
	}
	if _, err := client.ListStatuses(context.Background(), "objects", "people", "a1", false); err != nil {
		t.Fatalf("list statuses: %v", err)
	}
	if _, err := client.CreateStatus(context.Background(), "objects", "people", "a1", map[string]any{"title": "Open"}); err != nil {
		t.Fatalf("create status: %v", err)
	}
	if _, err := client.UpdateStatus(context.Background(), "objects", "people", "a1", "s1", map[string]any{"title": "Closed"}); err != nil {
		t.Fatalf("update status: %v", err)
	}
}
