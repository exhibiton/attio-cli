package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExecuteAttributesCommandMatrixJSON(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/attributes":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"attribute_id": "a1"}, "title": "Stage", "api_type": "status", "is_archived": false}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/people/attributes":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"attribute_id": "a1"}, "title": "Stage", "api_type": "status", "is_archived": false}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/attributes/stage":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"attribute_id": "a1"}, "title": "Stage", "api_type": "status", "is_archived": false}})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/objects/people/attributes/stage":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"attribute_id": "a1"}, "title": "Stage 2", "api_type": "status", "is_archived": false}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/attributes/stage/options":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"option_id": "o1"}, "title": "Open", "is_archived": false}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/people/attributes/stage/options":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"option_id": "o1"}, "title": "Open", "is_archived": false}})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/objects/people/attributes/stage/options/o1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"option_id": "o1"}, "title": "Closed", "is_archived": false}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/attributes/stage/statuses":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": map[string]any{"status_id": "s1"}, "title": "New", "is_archived": false}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/people/attributes/stage/statuses":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"status_id": "s1"}, "title": "New", "is_archived": false}})
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/objects/people/attributes/stage/statuses/s1":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": map[string]any{"status_id": "s1"}, "title": "Qualified", "is_archived": false}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{name: "list", args: []string{"--json", "attributes", "list", "objects", "people"}, contains: `"title": "Stage"`},
		{name: "create", args: []string{"--json", "attributes", "create", "objects", "people", "--data", `{"title":"Stage"}`}, contains: `"attribute_id": "a1"`},
		{name: "get", args: []string{"--json", "attributes", "get", "objects", "people", "stage"}, contains: `"api_type": "status"`},
		{name: "update", args: []string{"--json", "attributes", "update", "objects", "people", "stage", "--data", `{"title":"Stage 2"}`}, contains: `"Stage 2"`},
		{name: "options list", args: []string{"--json", "attributes", "options", "list", "objects", "people", "stage"}, contains: `"option_id": "o1"`},
		{name: "options create", args: []string{"--json", "attributes", "options", "create", "objects", "people", "stage", "--data", `{"title":"Open"}`}, contains: `"option_id": "o1"`},
		{name: "options update", args: []string{"--json", "attributes", "options", "update", "objects", "people", "stage", "o1", "--data", `{"title":"Closed"}`}, contains: `"Closed"`},
		{name: "statuses list", args: []string{"--json", "attributes", "statuses", "list", "objects", "people", "stage"}, contains: `"status_id": "s1"`},
		{name: "statuses create", args: []string{"--json", "attributes", "statuses", "create", "objects", "people", "stage", "--data", `{"title":"New"}`}, contains: `"status_id": "s1"`},
		{name: "statuses update", args: []string{"--json", "attributes", "statuses", "update", "objects", "people", "stage", "s1", "--data", `{"title":"Qualified"}`}, contains: `"Qualified"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := captureExecute(t, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
			}
			if strings.TrimSpace(stderr) != "" {
				t.Fatalf("expected empty stderr, got %q", stderr)
			}
			if !strings.Contains(stdout, tt.contains) {
				t.Fatalf("expected output to contain %q, got: %s", tt.contains, stdout)
			}
		})
	}
}

func TestExecuteAttributesInvalidTarget(t *testing.T) {
	setupCLIEnv(t)
	t.Setenv("ATTIO_API_KEY", "env-key")

	_, stderr, err := captureExecute(t, []string{"attributes", "list", "records", "people"})
	if err == nil {
		t.Fatalf("expected usage error")
	}
	if ExitCode(err) != ExitCodeUsage {
		t.Fatalf("expected usage exit code %d, got %d", ExitCodeUsage, ExitCode(err))
	}
	if !strings.Contains(stderr, "target must be 'objects' or 'lists'") {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}
