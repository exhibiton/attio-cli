package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReadRawInput(t *testing.T) {
	t.Run("inline", func(t *testing.T) {
		got, err := readRawInput(`{"ok":true}`)
		if err != nil {
			t.Fatalf("readRawInput inline: %v", err)
		}
		if string(got) != `{"ok":true}` {
			t.Fatalf("unexpected inline bytes: %q", string(got))
		}
	})

	t.Run("missing value", func(t *testing.T) {
		_, err := readRawInput("   ")
		if err == nil {
			t.Fatalf("expected usage error")
		}
		if ExitCode(err) != ExitCodeUsage {
			t.Fatalf("unexpected exit code: %d", ExitCode(err))
		}
	})

	t.Run("from file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "payload.json")
		if err := os.WriteFile(path, []byte(`{"a":1}`), 0o600); err != nil {
			t.Fatalf("write temp file: %v", err)
		}
		got, err := readRawInput("@" + path)
		if err != nil {
			t.Fatalf("readRawInput file: %v", err)
		}
		if string(got) != `{"a":1}` {
			t.Fatalf("unexpected file bytes: %q", string(got))
		}
	})

	t.Run("from stdin", func(t *testing.T) {
		orig := os.Stdin
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("pipe: %v", err)
		}
		_, _ = w.WriteString(`{"from":"stdin"}`)
		_ = w.Close()
		os.Stdin = r
		t.Cleanup(func() {
			os.Stdin = orig
			_ = r.Close()
		})

		got, err := readRawInput("-")
		if err != nil {
			t.Fatalf("readRawInput stdin: %v", err)
		}
		if string(got) != `{"from":"stdin"}` {
			t.Fatalf("unexpected stdin bytes: %q", string(got))
		}
	})
}

func TestReadJSONValueInput(t *testing.T) {
	v, err := readJSONValueInput(`[{"x":1}]`)
	if err != nil {
		t.Fatalf("readJSONValueInput: %v", err)
	}
	arr, ok := v.([]any)
	if !ok || len(arr) != 1 {
		t.Fatalf("unexpected value: %#v", v)
	}

	_, err = readJSONValueInput("not-json")
	if err == nil {
		t.Fatalf("expected invalid JSON usage error")
	}
	if ExitCode(err) != ExitCodeUsage {
		t.Fatalf("unexpected exit code: %d", ExitCode(err))
	}
}

func TestReadJSONObjectInput(t *testing.T) {
	obj, err := readJSONObjectInput(`{"a":1}`)
	if err != nil {
		t.Fatalf("readJSONObjectInput: %v", err)
	}
	if got := anyString(obj["a"]); got != "1" {
		t.Fatalf("unexpected object value: %q", got)
	}

	_, err = readJSONObjectInput(`[1,2]`)
	if err == nil {
		t.Fatalf("expected object validation error")
	}
	if ExitCode(err) != ExitCodeUsage {
		t.Fatalf("unexpected exit code: %d", ExitCode(err))
	}
}

func TestParseOptionalBoolFlag(t *testing.T) {
	v, err := parseOptionalBoolFlag("", "--flag")
	if err != nil {
		t.Fatalf("unexpected error for empty bool: %v", err)
	}
	if v != nil {
		t.Fatalf("expected nil for empty bool")
	}

	v, err = parseOptionalBoolFlag("true", "--flag")
	if err != nil {
		t.Fatalf("unexpected true parse error: %v", err)
	}
	if v == nil || *v != true {
		t.Fatalf("expected parsed true, got %#v", v)
	}

	_, err = parseOptionalBoolFlag("maybe", "--flag")
	if err == nil {
		t.Fatalf("expected invalid bool usage error")
	}
	if ExitCode(err) != ExitCodeUsage {
		t.Fatalf("unexpected exit code: %d", ExitCode(err))
	}
}

func TestParseJSONArrayFlag(t *testing.T) {
	items, err := parseJSONArrayFlag(`[1,{"a":2}]`, "--arr")
	if err != nil {
		t.Fatalf("parseJSONArrayFlag: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	items, err = parseJSONArrayFlag("", "--arr")
	if err != nil {
		t.Fatalf("empty parseJSONArrayFlag should not error: %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil for empty value")
	}

	_, err = parseJSONArrayFlag(`{"x":1}`, "--arr")
	if err == nil {
		t.Fatalf("expected usage error for non-array")
	}
	if ExitCode(err) != ExitCodeUsage {
		t.Fatalf("unexpected exit code: %d", ExitCode(err))
	}
}

func TestParseTaskAssigneesFlag(t *testing.T) {
	got, err := parseTaskAssigneesFlag("alice@example.com,member-1")
	if err != nil {
		t.Fatalf("parseTaskAssigneesFlag: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 assignees, got %d", len(got))
	}

	first, _ := got[0].(map[string]any)
	second, _ := got[1].(map[string]any)
	if first["workspace_member_email_address"] != "alice@example.com" {
		t.Fatalf("unexpected first assignee: %#v", first)
	}
	if second["referenced_actor_type"] != "workspace-member" || second["referenced_actor_id"] != "member-1" {
		t.Fatalf("unexpected second assignee: %#v", second)
	}
}

func TestHelpersValueExtraction(t *testing.T) {
	record := map[string]any{
		"values": map[string]any{
			"name":            []any{map[string]any{"full_name": "Ada Lovelace"}},
			"email_addresses": []any{map[string]any{"email_address": "ada@example.com"}},
		},
	}
	if got := recordValueSummary(record, "name"); got != "Ada Lovelace" {
		t.Fatalf("unexpected record name summary: %q", got)
	}
	if got := recordValueSummary(record, "email_addresses"); got != "ada@example.com" {
		t.Fatalf("unexpected record email summary: %q", got)
	}

	task := map[string]any{
		"is_completed": true,
		"assignees": []any{
			map[string]any{"workspace_member_email_address": "a@example.com"},
			map[string]any{"referenced_actor_id": "member-1"},
		},
	}
	if got := taskStatusSummary(task); got != "completed" {
		t.Fatalf("unexpected task status summary: %q", got)
	}
	if got := taskAssigneeSummary(task); got != "a@example.com,member-1" {
		t.Fatalf("unexpected task assignee summary: %q", got)
	}

	meeting := map[string]any{"participants": []any{map[string]any{"name": "Ada"}, map[string]any{"name": "Lin"}}}
	if got := meetingParticipantsSummary(meeting); got != "2" {
		t.Fatalf("unexpected participant summary: %q", got)
	}
	meeting["participant_count"] = "4"
	if got := meetingParticipantsSummary(meeting); got != "4" {
		t.Fatalf("unexpected participant_count override summary: %q", got)
	}
}

func TestMapAndTypeHelpers(t *testing.T) {
	if hasMapKey(nil, "x") {
		t.Fatalf("hasMapKey should be false for nil map")
	}
	if !hasMapKey(map[string]any{"x": 1}, "x") {
		t.Fatalf("hasMapKey should be true")
	}

	m := map[string]any{"value": []any{map[string]any{"name": "first"}}}
	if got := mapFromJSONField(m, "value"); got["name"] != "first" {
		t.Fatalf("unexpected mapFromJSONField value: %#v", got)
	}

	if got := anyString(true); got != "true" {
		t.Fatalf("unexpected bool anyString: %q", got)
	}
	if got := anyString(42.0); got != "42" {
		t.Fatalf("unexpected float anyString: %q", got)
	}
	if got := anyString(42.25); got != "42.25" {
		t.Fatalf("unexpected decimal anyString: %q", got)
	}

	if got := mapString(map[string]any{"v": 1}, "v"); got != "1" {
		t.Fatalf("unexpected mapString: %q", got)
	}
	if got := mapMap(map[string]any{"m": map[string]any{"x": "y"}}, "m"); !reflect.DeepEqual(got, map[string]any{"x": "y"}) {
		t.Fatalf("unexpected mapMap: %#v", got)
	}

	if got, ok := intFromAnyValue("12"); !ok || got != 12 {
		t.Fatalf("unexpected intFromAnyValue parse: got=%d ok=%v", got, ok)
	}
	if _, ok := intFromAnyValue("bad"); ok {
		t.Fatalf("expected intFromAnyValue to fail for bad input")
	}

	if got := idString(map[string]any{"record_id": "r1"}); got != "r1" {
		t.Fatalf("unexpected record idString: %q", got)
	}
	if got := idString(map[string]any{"entry_id": "e1"}); got != "e1" {
		t.Fatalf("unexpected entry idString: %q", got)
	}
	if got := idString("plain"); got != "plain" {
		t.Fatalf("unexpected plain idString: %q", got)
	}
}

func TestExpandPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := expandPath("~/data.json")
	if err != nil {
		t.Fatalf("expandPath: %v", err)
	}
	if !strings.HasPrefix(got, home) {
		t.Fatalf("expected expanded path to start with home %q, got %q", home, got)
	}
}

func TestMaskSecretEdgeCases(t *testing.T) {
	if got := maskSecret(""); got != "" {
		t.Fatalf("expected empty secret to stay empty, got %q", got)
	}
	if got := maskSecret("abcd1234"); got != "********" {
		t.Fatalf("expected short secret to be fully masked, got %q", got)
	}
	if got := maskSecret("abcd12345"); got != "*****2345" {
		t.Fatalf("expected long secret to preserve last 4 chars, got %q", got)
	}
}

func TestIntFromAnyValueBranches(t *testing.T) {
	if got, ok := intFromAnyValue(nil); ok || got != 0 {
		t.Fatalf("nil input should fail conversion, got=%d ok=%v", got, ok)
	}
	if got, ok := intFromAnyValue(int32(11)); !ok || got != 11 {
		t.Fatalf("int32 conversion failed, got=%d ok=%v", got, ok)
	}
	if got, ok := intFromAnyValue(int64(12)); !ok || got != 12 {
		t.Fatalf("int64 conversion failed, got=%d ok=%v", got, ok)
	}
	if got, ok := intFromAnyValue(float64(13.9)); !ok || got != 13 {
		t.Fatalf("float64 conversion failed, got=%d ok=%v", got, ok)
	}
	if got, ok := intFromAnyValue(json.Number("14")); !ok || got != 14 {
		t.Fatalf("json number conversion failed, got=%d ok=%v", got, ok)
	}
	if _, ok := intFromAnyValue(json.Number("14.5")); ok {
		t.Fatalf("expected non-integer json.Number conversion failure")
	}
	if _, ok := intFromAnyValue(struct{}{}); ok {
		t.Fatalf("expected unsupported type conversion failure")
	}
}

func TestMapFromJSONFieldBranches(t *testing.T) {
	if got := mapFromJSONField(nil, "x"); got != nil {
		t.Fatalf("expected nil map result for nil input, got %#v", got)
	}
	if got := mapFromJSONField(map[string]any{"x": map[string]any{"id": "one"}}, "x"); got["id"] != "one" {
		t.Fatalf("expected direct map value, got %#v", got)
	}
	if got := mapFromJSONField(map[string]any{"x": []any{}}, "x"); got != nil {
		t.Fatalf("expected nil for empty slice branch, got %#v", got)
	}
	if got := mapFromJSONField(map[string]any{"x": []any{"not-a-map"}}, "x"); got != nil {
		t.Fatalf("expected nil for slice-first-non-map branch, got %#v", got)
	}
	if got := mapFromJSONField(map[string]any{"x": "plain"}, "x"); got != nil {
		t.Fatalf("expected nil for default branch, got %#v", got)
	}
}

func TestStringOrSummaryFromValueBranches(t *testing.T) {
	if got := stringOrSummaryFromValue(nil); got != "" {
		t.Fatalf("expected nil summary to be empty, got %q", got)
	}
	if got := stringOrSummaryFromValue("abc"); got != "abc" {
		t.Fatalf("expected string summary passthrough, got %q", got)
	}
	if got := stringOrSummaryFromValue(map[string]any{"email_address": "ada@example.com"}); got != "ada@example.com" {
		t.Fatalf("expected email_address summary, got %q", got)
	}
	if got := stringOrSummaryFromValue(map[string]any{"unknown": "x"}); got != "{\"unknown\":\"x\"}" {
		t.Fatalf("expected map fallback JSON summary, got %q", got)
	}
	if got := stringOrSummaryFromValue([]any{nil, map[string]any{"title": "Pipeline Review"}}); got != "Pipeline Review" {
		t.Fatalf("expected first non-empty slice summary, got %q", got)
	}
	if got := stringOrSummaryFromValue([]any{nil, []any{}}); got != "" {
		t.Fatalf("expected empty slice summary when no values found, got %q", got)
	}
	if got := stringOrSummaryFromValue(true); got != "true" {
		t.Fatalf("expected default anyString path for bool, got %q", got)
	}
}
