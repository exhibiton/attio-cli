package outfmt

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestFromFlags(t *testing.T) {
	if _, err := FromFlags(true, true); err == nil {
		t.Fatalf("expected error for --json + --plain")
	}
	mode, err := FromFlags(true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mode.JSON || mode.Plain {
		t.Fatalf("unexpected mode: %+v", mode)
	}
}

func TestWriteJSONWithResultsOnly(t *testing.T) {
	ctx := context.Background()
	ctx = WithMode(ctx, Mode{JSON: true})
	ctx = WithJSONTransform(ctx, JSONTransform{ResultsOnly: true})

	in := map[string]any{"data": []map[string]any{{"id": 1}, {"id": 2}}, "meta": "x"}
	var buf bytes.Buffer
	if err := WriteJSON(ctx, &buf, in); err != nil {
		t.Fatalf("write json: %v", err)
	}

	if strings.Contains(buf.String(), "meta") {
		t.Fatalf("expected meta to be removed in results-only output: %s", buf.String())
	}

	var out []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out))
	}
}

func TestWriteJSONWithSelect(t *testing.T) {
	ctx := context.Background()
	ctx = WithMode(ctx, Mode{JSON: true})
	ctx = WithJSONTransform(ctx, JSONTransform{Select: []string{"workspace.name", "active"}})

	in := map[string]any{"active": true, "workspace": map[string]any{"name": "Failup", "slug": "failup"}}
	var buf bytes.Buffer
	if err := WriteJSON(ctx, &buf, in); err != nil {
		t.Fatalf("write json: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 selected fields, got %#v", out)
	}
	if out["workspace.name"] != "Failup" {
		t.Fatalf("unexpected selected workspace name: %#v", out)
	}
}
