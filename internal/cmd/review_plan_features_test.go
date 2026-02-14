package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

func TestAllowlistHelpers(t *testing.T) {
	if got := normalizeCommandPath(" records   get <object> <record-id> "); got != "records get" {
		t.Fatalf("unexpected normalized command path: %q", got)
	}

	if err := enforceCommandAllowlist("records get <object> <record-id>", "records"); err != nil {
		t.Fatalf("expected top-level allowlist to pass, got %v", err)
	}
	if err := enforceCommandAllowlist("records get <object> <record-id>", "records get"); err != nil {
		t.Fatalf("expected full-path allowlist to pass, got %v", err)
	}
	if err := enforceCommandAllowlist("records get <object> <record-id>", "tasks"); err == nil {
		t.Fatalf("expected allowlist error")
	} else if ExitCode(err) != ExitCodeUsage {
		t.Fatalf("expected usage exit code, got %d", ExitCode(err))
	}
}

func TestJSONRequestedFromArgsAndParseBoolArg(t *testing.T) {
	if !jsonRequestedFromArgs([]string{"--json", "version"}) {
		t.Fatalf("expected --json to be detected")
	}
	if !jsonRequestedFromArgs([]string{"-j", "version"}) {
		t.Fatalf("expected -j to be detected")
	}
	if !jsonRequestedFromArgs([]string{"--json=true", "version"}) {
		t.Fatalf("expected --json=true to be detected")
	}
	if jsonRequestedFromArgs([]string{"--json=true", "--plain", "version"}) {
		t.Fatalf("expected --plain to disable json-requested detection")
	}
	if jsonRequestedFromArgs([]string{"--json=bad", "version"}) {
		t.Fatalf("invalid bool should not mark json requested")
	}

	if got, ok := parseBoolArg("true"); !ok || !got {
		t.Fatalf("expected parseBoolArg(true) => (true, true), got (%v, %v)", got, ok)
	}
	if _, ok := parseBoolArg("not-bool"); ok {
		t.Fatalf("expected parseBoolArg invalid input to fail")
	}
}

func TestAutoJSONEnabled(t *testing.T) {
	t.Setenv("ATTIO_AUTO_JSON", "1")
	if !autoJSONEnabled() {
		t.Fatalf("expected ATTIO_AUTO_JSON=1 to enable auto-json")
	}

	t.Setenv("ATTIO_AUTO_JSON", "bad")
	if autoJSONEnabled() {
		t.Fatalf("expected invalid ATTIO_AUTO_JSON value to disable auto-json")
	}
}

func TestParseTimeout(t *testing.T) {
	if d, err := parseTimeout("45s"); err != nil || d.String() != "45s" {
		t.Fatalf("unexpected timeout parse result d=%v err=%v", d, err)
	}

	if _, err := parseTimeout(""); err == nil {
		t.Fatalf("expected empty timeout error")
	}
	if _, err := parseTimeout("-1s"); err == nil {
		t.Fatalf("expected negative timeout error")
	}
	if _, err := parseTimeout("garbage"); err == nil {
		t.Fatalf("expected invalid timeout error")
	}
}

func TestExecuteJSONErrorEnvelope(t *testing.T) {
	setupCLIEnv(t)

	t.Run("parse-error", func(t *testing.T) {
		_, stderr, err := captureExecute(t, []string{"--json", "records", "get"})
		if err == nil {
			t.Fatalf("expected parse error")
		}
		if ExitCode(err) != ExitCodeUsage {
			t.Fatalf("expected usage exit code %d, got %d", ExitCodeUsage, ExitCode(err))
		}

		var payload map[string]any
		if unmarshalErr := json.Unmarshal([]byte(stderr), &payload); unmarshalErr != nil {
			t.Fatalf("expected JSON error payload, got stderr=%q err=%v", stderr, unmarshalErr)
		}
		errObj, _ := payload["error"].(map[string]any)
		if errObj["kind"] != "usage" {
			t.Fatalf("expected usage kind in parse error payload: %#v", payload)
		}
	})

	t.Run("api-error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/records/missing" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"status_code":404,"type":"not_found_error","code":"not_found","message":"missing record"}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer srv.Close()

		t.Setenv("ATTIO_API_KEY", "env-key")
		t.Setenv("ATTIO_BASE_URL", srv.URL)

		stdout, stderr, err := captureExecute(t, []string{"--json", "records", "get", "people", "missing"})
		if err == nil {
			t.Fatalf("expected runtime error")
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("expected empty stdout on error, got %q", stdout)
		}

		var payload map[string]any
		if unmarshalErr := json.Unmarshal([]byte(stderr), &payload); unmarshalErr != nil {
			t.Fatalf("expected JSON error payload, got stderr=%q err=%v", stderr, unmarshalErr)
		}
		errObj, _ := payload["error"].(map[string]any)
		if errObj["status_code"] != float64(404) || errObj["code"] != "not_found" {
			t.Fatalf("unexpected API error payload: %#v", payload)
		}
	})
}

func TestExecuteEnableCommands(t *testing.T) {
	setupCLIEnv(t)

	_, stderr, err := captureExecute(t, []string{"--enable-commands", "version", "version"})
	if err != nil {
		t.Fatalf("expected allowed command to pass, err=%v stderr=%s", err, stderr)
	}

	_, stderr, err = captureExecute(t, []string{"--enable-commands", "version", "self"})
	if err == nil {
		t.Fatalf("expected blocked command error")
	}
	if ExitCode(err) != ExitCodeUsage {
		t.Fatalf("expected usage exit code %d, got %d", ExitCodeUsage, ExitCode(err))
	}
	if !strings.Contains(stderr, "not enabled") {
		t.Fatalf("expected allowlist message, got %q", stderr)
	}
}

func TestSchemaCommandAndBuilder(t *testing.T) {
	parser, _, err := newParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	root := buildCommandSchema(parser.Model.Node)
	if len(root.Commands) == 0 {
		t.Fatalf("expected root commands in schema builder output")
	}

	origStdout := os.Stdout
	tmp, err := os.CreateTemp(t.TempDir(), "schema-*.json")
	if err != nil {
		t.Fatalf("create temp schema file: %v", err)
	}
	os.Stdout = tmp
	t.Cleanup(func() {
		os.Stdout = origStdout
		_ = tmp.Close()
	})

	ctx := outfmt.WithMode(context.Background(), outfmt.Mode{JSON: true})
	if err := (&SchemaCmd{}).Run(ctx, parser); err != nil {
		t.Fatalf("schema run failed: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("close schema file: %v", err)
	}

	b, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatalf("read schema file: %v", err)
	}
	var payload map[string]any
	if unmarshalErr := json.Unmarshal(b, &payload); unmarshalErr != nil {
		t.Fatalf("schema output must be valid JSON: %v", unmarshalErr)
	}
	schemaRoot, ok := payload["root"].(map[string]any)
	if !ok {
		t.Fatalf("expected root schema object, got %#v", payload["root"])
	}
	if commands, ok := schemaRoot["commands"].([]any); !ok || len(commands) == 0 {
		t.Fatalf("expected commands in schema output, got %#v", schemaRoot["commands"])
	}
}

func TestExecuteIDOnlyAndPaginationMetadata(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2/objects/people/records":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"id": map[string]any{"record_id": "r1"}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/tasks":
			if r.URL.Query().Get("limit") != "1" || r.URL.Query().Get("offset") != "2" {
				t.Fatalf("unexpected tasks query: %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{"id": map[string]any{"task_id": "t1"}, "content": "Follow up"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--id-only", "records", "create", "people", "--data", `{"values":{"name":[{"full_name":"Ada"}]}}`})
	if err != nil {
		t.Fatalf("records create --id-only failed: %v stderr=%s", err, stderr)
	}
	if strings.TrimSpace(stdout) != "r1" {
		t.Fatalf("expected id-only output r1, got %q", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--json", "tasks", "list", "--limit", "1", "--offset", "2"})
	if err != nil {
		t.Fatalf("tasks list --json failed: %v stderr=%s", err, stderr)
	}
	var payload map[string]any
	if unmarshalErr := json.Unmarshal([]byte(stdout), &payload); unmarshalErr != nil {
		t.Fatalf("unmarshal tasks output: %v output=%s", unmarshalErr, stdout)
	}
	pagination, _ := payload["pagination"].(map[string]any)
	if pagination["limit"] != float64(1) || pagination["offset"] != float64(2) || pagination["has_more"] != true {
		t.Fatalf("unexpected pagination metadata: %#v", pagination)
	}
}

func TestExecuteAutoJSONWhenPiped(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/self" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"active":         true,
				"workspace_name": "Auto JSON",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	t.Setenv("ATTIO_AUTO_JSON", "1")
	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"self"})
	if err != nil {
		t.Fatalf("self with ATTIO_AUTO_JSON failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"workspace_name": "Auto JSON"`) {
		t.Fatalf("expected auto-json self output, got %s", stdout)
	}
}

func TestExecuteTablePathsAndNonJSONDryRun(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/lists":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
				"id":            map[string]any{"list_id": "l1"},
				"api_slug":      "prospects",
				"name":          "Prospects",
				"parent_object": map[string]any{"api_slug": "people"},
			}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/attributes":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
				"id":          map[string]any{"attribute_id": "a1"},
				"title":       "Stage",
				"api_type":    "status",
				"is_archived": false,
			}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/attributes/a1/options":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
				"id":          map[string]any{"option_id": "o1"},
				"title":       "Hot",
				"is_archived": false,
			}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/attributes/a1/statuses":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
				"id":          map[string]any{"status_id": "s1"},
				"title":       "Open",
				"is_archived": false,
			}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/lists/prospects/entries/e1/attributes/stage/values":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
				"active_from": "2025-01-01T00:00:00Z",
				"value":       "Open",
			}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects/people/records/r1/entries":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
				"id":         map[string]any{"entry_id": "e1"},
				"created_at": "2025-01-01T00:00:00Z",
				"web_url":    "https://app.attio.com/e/1",
			}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/objects":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"lists", "list"})
	if err != nil {
		t.Fatalf("lists list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "PARENT_OBJECT") || !strings.Contains(stdout, "prospects") {
		t.Fatalf("unexpected lists table output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"attributes", "list", "objects", "people"})
	if err != nil {
		t.Fatalf("attributes list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "TYPE") || !strings.Contains(stdout, "Stage") {
		t.Fatalf("unexpected attributes table output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"attributes", "options", "list", "objects", "people", "a1"})
	if err != nil {
		t.Fatalf("attributes options list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "IS_ARCHIVED") || !strings.Contains(stdout, "Hot") {
		t.Fatalf("unexpected options table output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"attributes", "statuses", "list", "objects", "people", "a1"})
	if err != nil {
		t.Fatalf("attributes statuses list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "Open") {
		t.Fatalf("unexpected statuses table output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"entries", "values", "list", "prospects", "e1", "stage"})
	if err != nil {
		t.Fatalf("entries values list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "VALUE") || !strings.Contains(stdout, "Open") {
		t.Fatalf("unexpected entries values table output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"records", "entries", "list", "people", "r1"})
	if err != nil {
		t.Fatalf("records entries list failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "ENTRY_ID") || !strings.Contains(stdout, "e1") {
		t.Fatalf("unexpected records entries table output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"--dry-run", "objects", "create", "--data", `{"api_slug":"dry_run_obj"}`})
	if err != nil {
		t.Fatalf("non-json dry-run failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "[dry-run] objects create") {
		t.Fatalf("unexpected dry-run non-json output: %s", stdout)
	}

	_, stderr, err = captureExecute(t, []string{"--fail-empty", "objects", "list"})
	if err == nil {
		t.Fatalf("expected fail-empty error")
	}
	if ExitCode(err) != ExitCodeNoResult {
		t.Fatalf("expected no-result exit code %d, got %d", ExitCodeNoResult, ExitCode(err))
	}
	if !strings.Contains(stderr, "No results") {
		t.Fatalf("expected non-json fail-empty stderr hint, got %q", stderr)
	}
}

func TestExecuteAuthLogoutNonJSON(t *testing.T) {
	setupCLIEnv(t)

	stdout, stderr, err := captureExecute(t, []string{"auth", "login", "--api-key", "real-secret-key"})
	if err != nil {
		t.Fatalf("auth login failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "Stored API key") {
		t.Fatalf("unexpected login output: %s", stdout)
	}

	stdout, stderr, err = captureExecute(t, []string{"auth", "logout"})
	if err != nil {
		t.Fatalf("auth logout failed: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "Removed API key") {
		t.Fatalf("unexpected logout output: %s", stdout)
	}
}
