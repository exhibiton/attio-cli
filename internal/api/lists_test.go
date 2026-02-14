package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListsAPI(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/lists":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"api_slug": "pipeline"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/lists":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"api_slug": "pipeline"}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/lists/pipeline":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"api_slug": "pipeline"}})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/lists/pipeline":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"api_slug": "pipeline", "name": "Pipeline"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient("test-key", srv.URL)

	lists, err := client.ListLists(context.Background())
	if err != nil {
		t.Fatalf("list lists: %v", err)
	}
	if len(lists) != 1 || lists[0]["api_slug"] != "pipeline" {
		t.Fatalf("unexpected lists payload: %#v", lists)
	}

	created, err := client.CreateList(context.Background(), map[string]any{"api_slug": "pipeline"})
	if err != nil {
		t.Fatalf("create list: %v", err)
	}
	if created["api_slug"] != "pipeline" {
		t.Fatalf("unexpected create list payload: %#v", created)
	}

	got, err := client.GetList(context.Background(), "pipeline")
	if err != nil {
		t.Fatalf("get list: %v", err)
	}
	if got["api_slug"] != "pipeline" {
		t.Fatalf("unexpected get list payload: %#v", got)
	}

	updated, err := client.UpdateList(context.Background(), "pipeline", map[string]any{"name": "Pipeline"})
	if err != nil {
		t.Fatalf("update list: %v", err)
	}
	if updated["name"] != "Pipeline" {
		t.Fatalf("unexpected update list payload: %#v", updated)
	}
}
