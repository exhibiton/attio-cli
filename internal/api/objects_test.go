package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestObjectsAPI(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"api_slug": "people"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if _, ok := body["data"].(map[string]any); !ok {
				t.Fatalf("expected wrapped data object")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"api_slug": "people"}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"api_slug": "people"}})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/objects/people":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"api_slug": "people", "name": "People"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient("test-key", srv.URL)

	objects, err := client.ListObjects(context.Background())
	if err != nil {
		t.Fatalf("list objects: %v", err)
	}
	if len(objects) != 1 || objects[0]["api_slug"] != "people" {
		t.Fatalf("unexpected objects payload: %#v", objects)
	}

	created, err := client.CreateObject(context.Background(), map[string]any{"api_slug": "people"})
	if err != nil {
		t.Fatalf("create object: %v", err)
	}
	if created["api_slug"] != "people" {
		t.Fatalf("unexpected create object payload: %#v", created)
	}

	got, err := client.GetObject(context.Background(), "people")
	if err != nil {
		t.Fatalf("get object: %v", err)
	}
	if got["api_slug"] != "people" {
		t.Fatalf("unexpected get object payload: %#v", got)
	}

	updated, err := client.UpdateObject(context.Background(), "people", map[string]any{"name": "People"})
	if err != nil {
		t.Fatalf("update object: %v", err)
	}
	if updated["name"] != "People" {
		t.Fatalf("unexpected update object payload: %#v", updated)
	}
}
