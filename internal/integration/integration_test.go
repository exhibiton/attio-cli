//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/failup-ventures/attio-cli/internal/api"
)

func integrationClient(t *testing.T) *api.Client {
	t.Helper()

	apiKey := strings.TrimSpace(os.Getenv("ATTIO_API_KEY"))
	if apiKey == "" {
		t.Skip("ATTIO_API_KEY not set")
	}
	baseURL := strings.TrimSpace(os.Getenv("ATTIO_BASE_URL"))
	return api.NewClient(apiKey, baseURL)
}

func integrationSkipEnabled(resource string) bool {
	resource = strings.ToLower(strings.TrimSpace(resource))
	if resource == "" {
		return false
	}
	raw := strings.TrimSpace(os.Getenv("ATTIO_IT_SKIP"))
	if raw == "" {
		return false
	}
	for _, part := range strings.Split(raw, ",") {
		if strings.ToLower(strings.TrimSpace(part)) == resource {
			return true
		}
	}
	return false
}

func requireIntegrationResource(t *testing.T, resource string) {
	t.Helper()
	if integrationSkipEnabled(resource) {
		t.Skipf("%s skipped via ATTIO_IT_SKIP", resource)
	}
}

func skipOnInsufficientScope(t *testing.T, resource string, err error) bool {
	t.Helper()
	if err == nil {
		return false
	}
	var ae *api.AttioError
	if !errors.As(err, &ae) {
		return false
	}
	if ae.StatusCode == 401 || ae.StatusCode == 403 {
		t.Skipf("skipping %s integration test due to scope: %v", resource, err)
		return true
	}
	return false
}

func idFromEntity(t *testing.T, entity map[string]any, key string) string {
	t.Helper()
	idObj, _ := entity["id"].(map[string]any)
	if idObj != nil {
		if id, _ := idObj[key].(string); strings.TrimSpace(id) != "" {
			return id
		}
	}
	if id, _ := entity[key].(string); strings.TrimSpace(id) != "" {
		return id
	}
	t.Fatalf("missing %q in entity: %#v", key, entity)
	return ""
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "client.timeout")
}

func retryDelete(fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		err := fn()
		if err == nil || api.IsNotFound(err) {
			return nil
		}
		if !isTimeoutError(err) {
			return err
		}
		lastErr = err
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}
	return lastErr
}

func createIntegrationPersonRecord(t *testing.T, ctx context.Context, client *api.Client) string {
	t.Helper()
	requireIntegrationResource(t, "records")

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	record, err := client.CreateRecord(ctx, "people", map[string]any{
		"values": map[string]any{
			"name": []any{"Integration CLI " + suffix},
			"email_addresses": []any{
				map[string]any{
					"email_address": fmt.Sprintf("integration.helper.%s@example.com", suffix),
				},
			},
		},
	})
	if skipOnInsufficientScope(t, "records", err) {
		return ""
	}
	if err != nil {
		t.Fatalf("create record: %v", err)
	}
	recordID := idFromEntity(t, record, "record_id")
	t.Cleanup(func() {
		if delErr := retryDelete(func() error { return client.DeleteRecord(ctx, "people", recordID) }); delErr != nil {
			t.Fatalf("cleanup delete record %s: %v", recordID, delErr)
		}
	})
	return recordID
}

func TestSelf(t *testing.T) {
	client := integrationClient(t)
	self, err := client.GetSelf(context.Background())
	if err != nil {
		t.Fatalf("get self: %v", err)
	}
	if !self.Active {
		t.Fatalf("expected active token")
	}
}

func TestObjectsList(t *testing.T) {
	client := integrationClient(t)
	objects, err := client.ListObjects(context.Background())
	if err != nil {
		t.Fatalf("list objects: %v", err)
	}
	if len(objects) == 0 {
		t.Fatalf("expected at least one object")
	}
}

func TestListsAndEntriesReadPaths(t *testing.T) {
	requireIntegrationResource(t, "lists")
	requireIntegrationResource(t, "entries")

	client := integrationClient(t)
	ctx := context.Background()

	lists, err := client.ListLists(ctx)
	if skipOnInsufficientScope(t, "lists", err) {
		return
	}
	if err != nil {
		t.Fatalf("list lists: %v", err)
	}
	if len(lists) == 0 {
		t.Skip("workspace has no lists")
	}

	listID := idFromEntity(t, lists[0], "list_id")
	_, err = client.GetList(ctx, listID)
	if skipOnInsufficientScope(t, "lists", err) {
		return
	}
	if err != nil {
		t.Fatalf("get list: %v", err)
	}

	_, err = client.QueryEntries(ctx, listID, nil, nil, 1, 0)
	if skipOnInsufficientScope(t, "entries", err) {
		return
	}
	if err != nil {
		t.Fatalf("query entries: %v", err)
	}
}

func TestMeetingsListReadPath(t *testing.T) {
	requireIntegrationResource(t, "meetings")

	client := integrationClient(t)
	ctx := context.Background()

	_, _, err := client.ListMeetings(ctx, 1, "", "", "", "", "", "", "", "")
	if skipOnInsufficientScope(t, "meetings", err) {
		return
	}
	if err != nil {
		t.Fatalf("list meetings: %v", err)
	}
}

func TestRecordsCRUD(t *testing.T) {
	requireIntegrationResource(t, "records")

	client := integrationClient(t)
	ctx := context.Background()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	record, err := client.CreateRecord(ctx, "people", map[string]any{
		"values": map[string]any{
			"name": []any{"Integration Records " + suffix},
			"email_addresses": []any{
				map[string]any{
					"email_address": fmt.Sprintf("integration.records.%s@example.com", suffix),
				},
			},
		},
	})
	if skipOnInsufficientScope(t, "records", err) {
		return
	}
	if err != nil {
		t.Fatalf("create record: %v", err)
	}

	recordID := idFromEntity(t, record, "record_id")
	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}
		if delErr := retryDelete(func() error { return client.DeleteRecord(ctx, "people", recordID) }); delErr != nil {
			t.Fatalf("cleanup delete record %s: %v", recordID, delErr)
		}
	})

	got, err := client.GetRecord(ctx, "people", recordID)
	if err != nil {
		t.Fatalf("get record: %v", err)
	}
	if idFromEntity(t, got, "record_id") != recordID {
		t.Fatalf("unexpected record id after get")
	}

	_, err = client.UpdateRecord(ctx, "people", recordID, map[string]any{
		"values": map[string]any{
			"name": []any{"Integration Updated " + suffix},
		},
	})
	if err != nil {
		t.Fatalf("update record: %v", err)
	}

	if err := retryDelete(func() error { return client.DeleteRecord(ctx, "people", recordID) }); err != nil {
		t.Fatalf("delete record: %v", err)
	}
	deleted = true
}

func TestTasksCRUD(t *testing.T) {
	requireIntegrationResource(t, "tasks")

	client := integrationClient(t)
	ctx := context.Background()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	task, err := client.CreateTask(ctx, map[string]any{
		"content":        "Integration task " + suffix,
		"format":         "plaintext",
		"deadline_at":    nil,
		"is_completed":   false,
		"linked_records": []any{},
		"assignees":      []any{},
	})
	if skipOnInsufficientScope(t, "tasks", err) {
		return
	}
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	taskID := idFromEntity(t, task, "task_id")
	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}
		if delErr := retryDelete(func() error { return client.DeleteTask(ctx, taskID) }); delErr != nil {
			t.Fatalf("cleanup delete task %s: %v", taskID, delErr)
		}
	})

	got, err := client.GetTask(ctx, taskID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if idFromEntity(t, got, "task_id") != taskID {
		t.Fatalf("unexpected task id after get")
	}

	_, err = client.UpdateTask(ctx, taskID, map[string]any{"is_completed": true})
	if err != nil {
		t.Fatalf("update task: %v", err)
	}

	if err := retryDelete(func() error { return client.DeleteTask(ctx, taskID) }); err != nil {
		t.Fatalf("delete task: %v", err)
	}
	deleted = true
}

func TestNotesCRUD(t *testing.T) {
	requireIntegrationResource(t, "notes")

	client := integrationClient(t)
	ctx := context.Background()

	recordID := createIntegrationPersonRecord(t, ctx, client)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	note, err := client.CreateNote(ctx, map[string]any{
		"parent_object":    "people",
		"parent_record_id": recordID,
		"title":            "Integration note " + suffix,
		"format":           "plaintext",
		"content":          "integration note body",
	})
	if skipOnInsufficientScope(t, "notes", err) {
		return
	}
	if err != nil {
		t.Fatalf("create note: %v", err)
	}

	noteID := idFromEntity(t, note, "note_id")
	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}
		if delErr := retryDelete(func() error { return client.DeleteNote(ctx, noteID) }); delErr != nil {
			t.Fatalf("cleanup delete note %s: %v", noteID, delErr)
		}
	})

	got, err := client.GetNote(ctx, noteID)
	if err != nil {
		t.Fatalf("get note: %v", err)
	}
	if idFromEntity(t, got, "note_id") != noteID {
		t.Fatalf("unexpected note id after get")
	}

	if err := retryDelete(func() error { return client.DeleteNote(ctx, noteID) }); err != nil {
		t.Fatalf("delete note: %v", err)
	}
	deleted = true
}

func TestCommentsCRUD(t *testing.T) {
	requireIntegrationResource(t, "comments")

	client := integrationClient(t)
	ctx := context.Background()

	recordID := createIntegrationPersonRecord(t, ctx, client)
	members, err := client.ListMembers(ctx)
	if skipOnInsufficientScope(t, "members", err) {
		return
	}
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(members) == 0 {
		t.Skip("workspace has no members")
	}
	authorID := idFromEntity(t, members[0], "workspace_member_id")

	comment, err := client.CreateComment(ctx, map[string]any{
		"format":  "plaintext",
		"content": "integration comment",
		"author": map[string]any{
			"type": "workspace-member",
			"id":   authorID,
		},
		"record": map[string]any{
			"object":    "people",
			"record_id": recordID,
		},
	})
	if skipOnInsufficientScope(t, "comments", err) {
		return
	}
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	commentID := idFromEntity(t, comment, "comment_id")
	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}
		if delErr := retryDelete(func() error { return client.DeleteComment(ctx, commentID) }); delErr != nil {
			t.Fatalf("cleanup delete comment %s: %v", commentID, delErr)
		}
	})

	got, err := client.GetComment(ctx, commentID)
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}
	if idFromEntity(t, got, "comment_id") != commentID {
		t.Fatalf("unexpected comment id after get")
	}

	if err := retryDelete(func() error { return client.DeleteComment(ctx, commentID) }); err != nil {
		t.Fatalf("delete comment: %v", err)
	}
	deleted = true
}

func TestWebhooksCRUD(t *testing.T) {
	requireIntegrationResource(t, "webhooks")

	client := integrationClient(t)
	ctx := context.Background()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	webhook, err := client.CreateWebhook(ctx, map[string]any{
		"target_url": fmt.Sprintf("https://example.com/webhooks/attio-cli/%s", suffix),
		"subscriptions": []any{
			map[string]any{
				"event_type": "task.created",
				"filter":     nil,
			},
		},
	})
	if skipOnInsufficientScope(t, "webhooks", err) {
		return
	}
	if err != nil {
		t.Fatalf("create webhook: %v", err)
	}

	webhookID := idFromEntity(t, webhook, "webhook_id")
	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}
		if delErr := retryDelete(func() error { return client.DeleteWebhook(ctx, webhookID) }); delErr != nil {
			t.Fatalf("cleanup delete webhook %s: %v", webhookID, delErr)
		}
	})

	got, err := client.GetWebhook(ctx, webhookID)
	if err != nil {
		t.Fatalf("get webhook: %v", err)
	}
	if idFromEntity(t, got, "webhook_id") != webhookID {
		t.Fatalf("unexpected webhook id after get")
	}

	_, err = client.UpdateWebhook(ctx, webhookID, map[string]any{
		"target_url": fmt.Sprintf("https://example.com/webhooks/attio-cli/%s/updated", suffix),
		"subscriptions": []any{
			map[string]any{
				"event_type": "task.updated",
				"filter":     nil,
			},
		},
	})
	if err != nil {
		t.Fatalf("update webhook: %v", err)
	}

	if err := retryDelete(func() error { return client.DeleteWebhook(ctx, webhookID) }); err != nil {
		t.Fatalf("delete webhook: %v", err)
	}
	deleted = true
}
