package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEntriesAPI(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2/lists/pipeline/entries":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			data, _ := body["data"].(map[string]any)
			if data["parent_record_id"] != "r1" {
				t.Fatalf("expected parent_record_id=r1, got %#v", data["parent_record_id"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"id": map[string]any{"entry_id": "e1"}, "web_url": "https://app.attio.com/e/1"},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/v2/lists/pipeline/entries":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			data, _ := body["data"].(map[string]any)
			if data["parent_record_id"] != "r1" {
				t.Fatalf("expected parent_record_id=r1, got %#v", data["parent_record_id"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"entry_id": "e1"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/lists/pipeline/entries/query":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			filter, _ := body["filter"].(map[string]any)
			if filter["stage"] != "Qualified" {
				t.Fatalf("expected filter stage Qualified, got %#v", filter["stage"])
			}
			sorts, _ := body["sorts"].([]any)
			if len(sorts) != 1 {
				t.Fatalf("expected 1 sort, got %d", len(sorts))
			}
			if body["limit"] != float64(2) {
				t.Fatalf("expected limit=2, got %#v", body["limit"])
			}
			if body["offset"] != float64(4) {
				t.Fatalf("expected offset=4, got %#v", body["offset"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"entry_id": "e1"}}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/lists/pipeline/entries/e1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"id": map[string]any{"entry_id": "e1"}, "web_url": "https://app.attio.com/e/1"},
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/lists/pipeline/entries/e1":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			data, _ := body["data"].(map[string]any)
			if data["stage"] != "updated" {
				t.Fatalf("expected stage=updated, got %#v", data["stage"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"id": map[string]any{"entry_id": "e1"}, "web_url": "https://app.attio.com/e/1-updated"},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/v2/lists/pipeline/entries/e1":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			data, _ := body["data"].(map[string]any)
			if data["stage"] != "replaced" {
				t.Fatalf("expected stage=replaced, got %#v", data["stage"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"id": map[string]any{"entry_id": "e1"}, "web_url": "https://app.attio.com/e/1-replaced"},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/lists/pipeline/entries/e1/attributes/stage/values":
			if r.URL.Query().Get("show_historic") != "true" {
				t.Fatalf("expected show_historic=true")
			}
			if r.URL.Query().Get("limit") != "3" {
				t.Fatalf("expected limit=3")
			}
			if r.URL.Query().Get("offset") != "1" {
				t.Fatalf("expected offset=1")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"value": "Qualified"}}})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/lists/pipeline/entries/e1":
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient("test-key", srv.URL)

	created, err := client.CreateEntry(context.Background(), "pipeline", map[string]any{"parent_record_id": "r1"})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if id := map[string]any(created["id"].(map[string]any))["entry_id"]; id != "e1" {
		t.Fatalf("unexpected create response: %#v", created)
	}

	asserted, err := client.AssertEntry(context.Background(), "pipeline", map[string]any{"parent_record_id": "r1"})
	if err != nil {
		t.Fatalf("assert entry: %v", err)
	}
	if id := map[string]any(asserted["id"].(map[string]any))["entry_id"]; id != "e1" {
		t.Fatalf("unexpected assert response: %#v", asserted)
	}

	entries, err := client.QueryEntries(
		context.Background(),
		"pipeline",
		map[string]any{"stage": "Qualified"},
		[]map[string]any{{"attribute": "created_at", "direction": "asc"}},
		2,
		4,
	)
	if err != nil {
		t.Fatalf("query entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry, err := client.GetEntry(context.Background(), "pipeline", "e1")
	if err != nil {
		t.Fatalf("get entry: %v", err)
	}
	if id := map[string]any(entry["id"].(map[string]any))["entry_id"]; id != "e1" {
		t.Fatalf("unexpected get response: %#v", entry)
	}

	updated, err := client.UpdateEntry(context.Background(), "pipeline", "e1", map[string]any{"stage": "updated"})
	if err != nil {
		t.Fatalf("update entry: %v", err)
	}
	if updated["web_url"] != "https://app.attio.com/e/1-updated" {
		t.Fatalf("unexpected update response: %#v", updated)
	}

	replaced, err := client.ReplaceEntry(context.Background(), "pipeline", "e1", map[string]any{"stage": "replaced"})
	if err != nil {
		t.Fatalf("replace entry: %v", err)
	}
	if replaced["web_url"] != "https://app.attio.com/e/1-replaced" {
		t.Fatalf("unexpected replace response: %#v", replaced)
	}

	values, err := client.ListEntryAttributeValues(context.Background(), "pipeline", "e1", "stage", true, 3, 1)
	if err != nil {
		t.Fatalf("list entry values: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}

	if err := client.DeleteEntry(context.Background(), "pipeline", "e1"); err != nil {
		t.Fatalf("delete entry: %v", err)
	}
}
