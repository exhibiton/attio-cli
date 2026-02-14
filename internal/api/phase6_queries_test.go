package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListTasksWithAllFilters(t *testing.T) {
	t.Parallel()

	isCompleted := true
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/tasks" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		if q.Get("limit") != "25" ||
			q.Get("offset") != "5" ||
			q.Get("sort") != "deadline_at:asc" ||
			q.Get("linked_object") != "people" ||
			q.Get("linked_record_id") != "rec_1" ||
			q.Get("assignee") != "member_1" ||
			q.Get("is_completed") != "true" {
			t.Fatalf("unexpected query params: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"task_id": "t1"}}}})
	}))
	defer srv.Close()

	client := NewClient("test-key", srv.URL)
	tasks, err := client.ListTasks(context.Background(), 25, 5, "deadline_at:asc", "people", "rec_1", "member_1", &isCompleted)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
}

func TestMeetingsQueryParamsAndCursorBranches(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings":
			q := r.URL.Query()
			if q.Get("limit") != "10" ||
				q.Get("cursor") != "cur_1" ||
				q.Get("sort") != "starts_at:desc" ||
				q.Get("participants") != "member_1" ||
				q.Get("linked_object") != "people" ||
				q.Get("linked_record_id") != "rec_1" ||
				q.Get("ends_from") != "2025-01-01T00:00:00Z" ||
				q.Get("starts_before") != "2025-01-31T00:00:00Z" ||
				q.Get("timezone") != "UTC" {
				t.Fatalf("unexpected meetings query params: %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":       []map[string]any{{"id": map[string]any{"meeting_id": "m1"}}},
				"pagination": map[string]any{"next_cursor": nil},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings/m1/call_recordings":
			q := r.URL.Query()
			if q.Get("limit") != "50" || q.Get("cursor") != "c1" {
				t.Fatalf("unexpected recordings query params: %s", r.URL.RawQuery)
			}
			next := "c2"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":       []map[string]any{{"id": map[string]any{"call_recording_id": "r1"}}},
				"pagination": map[string]any{"next_cursor": next},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/meetings/m1/call_recordings/r1/transcript":
			if r.URL.Query().Get("cursor") != "c2" {
				t.Fatalf("unexpected transcript query params: %s", r.URL.RawQuery)
			}
			next := "c3"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":       []map[string]any{{"text": "hello"}},
				"pagination": map[string]any{"next_cursor": next},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient("test-key", srv.URL)

	meetings, next, err := client.ListMeetings(
		context.Background(),
		10,
		"cur_1",
		"starts_at:desc",
		"member_1",
		"people",
		"rec_1",
		"2025-01-01T00:00:00Z",
		"2025-01-31T00:00:00Z",
		"UTC",
	)
	if err != nil {
		t.Fatalf("list meetings: %v", err)
	}
	if len(meetings) != 1 || next != "" {
		t.Fatalf("unexpected meetings response len=%d next=%q", len(meetings), next)
	}

	recordings, recordingsNext, err := client.ListCallRecordings(context.Background(), "m1", 50, "c1")
	if err != nil {
		t.Fatalf("list call recordings: %v", err)
	}
	if len(recordings) != 1 || recordingsNext != "c2" {
		t.Fatalf("unexpected recordings response len=%d next=%q", len(recordings), recordingsNext)
	}

	transcript, transcriptNext, err := client.GetTranscript(context.Background(), "m1", "r1", "c2")
	if err != nil {
		t.Fatalf("get transcript: %v", err)
	}
	if len(transcript) != 1 || transcriptNext != "c3" {
		t.Fatalf("unexpected transcript response len=%d next=%q", len(transcript), transcriptNext)
	}
}
