package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecordsAPI(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/people/records":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			data, _ := body["data"].(map[string]any)
			values, _ := data["values"].(map[string]any)
			if values["name"] == nil {
				t.Fatalf("expected values.name in create payload")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"id": map[string]any{"record_id": "r1"}, "web_url": "https://app.attio.com/r/1"},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/v2/objects/people/records":
			if got := r.URL.Query().Get("matching_attribute"); got != "email_addresses" {
				t.Fatalf("expected matching_attribute=email_addresses, got %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"record_id": "r1"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/people/records/query":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
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
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"record_id": "r1"}}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/records/search":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["query"] != "Ada" {
				t.Fatalf("expected search query Ada, got %#v", body["query"])
			}
			if body["limit"] != float64(5) {
				t.Fatalf("expected limit=5, got %#v", body["limit"])
			}
			objects, _ := body["objects"].([]any)
			if len(objects) != 1 || objects[0] != "people" {
				t.Fatalf("expected objects [people], got %#v", body["objects"])
			}
			requestAs, _ := body["request_as"].(map[string]any)
			if requestAs["type"] == nil {
				t.Fatalf("expected request_as.type in search payload")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{"id": map[string]any{"record_id": "r1"}}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/records/r1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"id": map[string]any{"record_id": "r1"}, "web_url": "https://app.attio.com/r/1"},
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/objects/people/records/r1":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			data, _ := body["data"].(map[string]any)
			values, _ := data["values"].(map[string]any)
			if values["name"] == nil {
				t.Fatalf("expected values.name in update payload")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"id": map[string]any{"record_id": "r1"}, "web_url": "https://app.attio.com/r/1-updated"},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/v2/objects/people/records/r1":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			data, _ := body["data"].(map[string]any)
			values, _ := data["values"].(map[string]any)
			if values["name"] == nil {
				t.Fatalf("expected values.name in replace payload")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"id": map[string]any{"record_id": "r1"}, "web_url": "https://app.attio.com/r/1-replaced"},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/records/r1/attributes/email_addresses/values":
			if r.URL.Query().Get("show_historic") != "true" {
				t.Fatalf("expected show_historic=true")
			}
			if r.URL.Query().Get("limit") != "3" {
				t.Fatalf("expected limit=3")
			}
			if r.URL.Query().Get("offset") != "1" {
				t.Fatalf("expected offset=1")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"value": "a@example.com"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/records/r1/entries":
			if r.URL.Query().Get("limit") != "2" {
				t.Fatalf("expected limit=2")
			}
			if r.URL.Query().Get("offset") != "3" {
				t.Fatalf("expected offset=3")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{"id": map[string]any{"entry_id": "e1"}}},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/objects/people/records/r1":
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient("test-key", srv.URL)

	created, err := client.CreateRecord(context.Background(), "people", map[string]any{"values": map[string]any{"name": []string{"Ada"}}})
	if err != nil {
		t.Fatalf("create record: %v", err)
	}
	if id := map[string]any(created["id"].(map[string]any))["record_id"]; id != "r1" {
		t.Fatalf("unexpected create response: %#v", created)
	}

	asserted, err := client.AssertRecord(context.Background(), "people", "email_addresses", map[string]any{"values": map[string]any{"name": []string{"Ada"}}})
	if err != nil {
		t.Fatalf("assert record: %v", err)
	}
	if id := map[string]any(asserted["id"].(map[string]any))["record_id"]; id != "r1" {
		t.Fatalf("unexpected assert response: %#v", asserted)
	}

	records, err := client.QueryRecords(
		context.Background(),
		"people",
		map[string]any{"name": "Ada"},
		[]map[string]any{{"attribute": "name", "direction": "asc"}},
		2,
		4,
	)
	if err != nil {
		t.Fatalf("query records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	searchDefault, err := client.SearchRecords(context.Background(), "Ada", 5, []string{"people"}, nil)
	if err != nil {
		t.Fatalf("search records default request_as: %v", err)
	}
	if len(searchDefault) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(searchDefault))
	}

	searchUser, err := client.SearchRecords(context.Background(), "Ada", 5, []string{"people"}, map[string]any{"type": "user", "id": "u1"})
	if err != nil {
		t.Fatalf("search records custom request_as: %v", err)
	}
	if len(searchUser) != 1 {
		t.Fatalf("expected 1 search result with custom request_as, got %d", len(searchUser))
	}

	record, err := client.GetRecord(context.Background(), "people", "r1")
	if err != nil {
		t.Fatalf("get record: %v", err)
	}
	if id := map[string]any(record["id"].(map[string]any))["record_id"]; id != "r1" {
		t.Fatalf("unexpected get response: %#v", record)
	}

	updated, err := client.UpdateRecord(context.Background(), "people", "r1", map[string]any{"values": map[string]any{"name": []string{"Ada Lovelace"}}})
	if err != nil {
		t.Fatalf("update record: %v", err)
	}
	if updated["web_url"] != "https://app.attio.com/r/1-updated" {
		t.Fatalf("unexpected update response: %#v", updated)
	}

	replaced, err := client.ReplaceRecord(context.Background(), "people", "r1", map[string]any{"values": map[string]any{"name": []string{"Ada Byron"}}})
	if err != nil {
		t.Fatalf("replace record: %v", err)
	}
	if replaced["web_url"] != "https://app.attio.com/r/1-replaced" {
		t.Fatalf("unexpected replace response: %#v", replaced)
	}

	values, err := client.ListRecordAttributeValues(context.Background(), "people", "r1", "email_addresses", true, 3, 1)
	if err != nil {
		t.Fatalf("list record values: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}

	entries, err := client.ListRecordEntries(context.Background(), "people", "r1", 2, 3)
	if err != nil {
		t.Fatalf("list record entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 record entry, got %d", len(entries))
	}

	if err := client.DeleteRecord(context.Background(), "people", "r1"); err != nil {
		t.Fatalf("delete record: %v", err)
	}
}
