package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPhase3ResourcesAPI(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/notes":
			if r.URL.Query().Get("parent_object") != "people" {
				t.Fatalf("expected parent_object filter")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"note_id": "n1"}, "title": "Note"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/notes":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"note_id": "n1"}, "title": "Note"}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/notes/n1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"note_id": "n1"}, "title": "Note"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/notes/n1":
			_, _ = w.Write([]byte(`{}`))

		case r.Method == http.MethodGet && r.URL.Path == "/v2/tasks":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"task_id": "t1"}, "content": "Task"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/tasks":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"task_id": "t1"}, "content": "Task"}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/t1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"task_id": "t1"}, "content": "Task"}})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/tasks/t1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"task_id": "t1"}, "content": "Task Updated"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/tasks/t1":
			_, _ = w.Write([]byte(`{}`))

		case r.Method == http.MethodPost && r.URL.Path == "/v2/comments":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"comment_id": "c1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/comments/c1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"comment_id": "c1"}}})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/comments/c1":
			_, _ = w.Write([]byte(`{}`))

		case r.Method == http.MethodGet && r.URL.Path == "/v2/threads":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"thread_id": "th1"}}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/threads/th1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"thread_id": "th1"}}})

		case r.Method == http.MethodGet && r.URL.Path == "/v2/webhooks":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"webhook_id": "w1"}}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/webhooks":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"webhook_id": "w1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/webhooks/w1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"webhook_id": "w1"}}})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/webhooks/w1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"webhook_id": "w1"}, "status": "active"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/webhooks/w1":
			_, _ = w.Write([]byte(`{}`))

		case r.Method == http.MethodGet && r.URL.Path == "/v2/workspace_members":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"workspace_member_id": "m1"}, "email_address": "m1@example.com"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/workspace_members/m1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"workspace_member_id": "m1"}, "email_address": "m1@example.com"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient("test-key", srv.URL)

	notes, err := client.ListNotes(context.Background(), "people", "r1", 10, 0)
	if err != nil || len(notes) != 1 {
		t.Fatalf("list notes failed: %v len=%d", err, len(notes))
	}
	if _, err := client.CreateNote(context.Background(), map[string]any{"title": "Note"}); err != nil {
		t.Fatalf("create note: %v", err)
	}
	if _, err := client.GetNote(context.Background(), "n1"); err != nil {
		t.Fatalf("get note: %v", err)
	}
	if err := client.DeleteNote(context.Background(), "n1"); err != nil {
		t.Fatalf("delete note: %v", err)
	}

	tasks, err := client.ListTasks(context.Background(), 20, 0, "", "", "", "", nil)
	if err != nil || len(tasks) != 1 {
		t.Fatalf("list tasks failed: %v len=%d", err, len(tasks))
	}
	if _, err := client.CreateTask(context.Background(), map[string]any{"content": "Task"}); err != nil {
		t.Fatalf("create task: %v", err)
	}
	if _, err := client.GetTask(context.Background(), "t1"); err != nil {
		t.Fatalf("get task: %v", err)
	}
	if _, err := client.UpdateTask(context.Background(), "t1", map[string]any{"is_completed": true}); err != nil {
		t.Fatalf("update task: %v", err)
	}
	if err := client.DeleteTask(context.Background(), "t1"); err != nil {
		t.Fatalf("delete task: %v", err)
	}

	if _, err := client.CreateComment(context.Background(), map[string]any{"body": "hi"}); err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if _, err := client.GetComment(context.Background(), "c1"); err != nil {
		t.Fatalf("get comment: %v", err)
	}
	if err := client.DeleteComment(context.Background(), "c1"); err != nil {
		t.Fatalf("delete comment: %v", err)
	}

	threads, err := client.ListThreads(context.Background(), "people", "r1", "", "", 10, 0)
	if err != nil || len(threads) != 1 {
		t.Fatalf("list threads failed: %v len=%d", err, len(threads))
	}
	if _, err := client.GetThread(context.Background(), "th1"); err != nil {
		t.Fatalf("get thread: %v", err)
	}

	webhooks, err := client.ListWebhooks(context.Background(), 10, 0)
	if err != nil || len(webhooks) != 1 {
		t.Fatalf("list webhooks failed: %v len=%d", err, len(webhooks))
	}
	if _, err := client.CreateWebhook(context.Background(), map[string]any{"target_url": "https://example.com"}); err != nil {
		t.Fatalf("create webhook: %v", err)
	}
	if _, err := client.GetWebhook(context.Background(), "w1"); err != nil {
		t.Fatalf("get webhook: %v", err)
	}
	if _, err := client.UpdateWebhook(context.Background(), "w1", map[string]any{"target_url": "https://example.com"}); err != nil {
		t.Fatalf("update webhook: %v", err)
	}
	if err := client.DeleteWebhook(context.Background(), "w1"); err != nil {
		t.Fatalf("delete webhook: %v", err)
	}

	members, err := client.ListMembers(context.Background())
	if err != nil || len(members) != 1 {
		t.Fatalf("list members failed: %v len=%d", err, len(members))
	}
	if _, err := client.GetMember(context.Background(), "m1"); err != nil {
		t.Fatalf("get member: %v", err)
	}
}
