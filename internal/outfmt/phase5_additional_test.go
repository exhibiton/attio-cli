package outfmt

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

type failingWriter struct{}

func (f failingWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestParseErrorAndModeContext(t *testing.T) {
	if got := (&ParseError{msg: "invalid mode"}).Error(); got != "invalid mode" {
		t.Fatalf("unexpected parse error message: %q", got)
	}

	if mode := FromContext(context.Background()); mode.JSON || mode.Plain {
		t.Fatalf("expected zero mode from empty context, got %+v", mode)
	}

	ctxWrong := context.WithValue(context.Background(), modeKey{}, "not-a-mode")
	if mode := FromContext(ctxWrong); mode.JSON || mode.Plain {
		t.Fatalf("expected zero mode from wrong context value, got %+v", mode)
	}

	ctx := WithMode(context.Background(), Mode{JSON: true})
	if !IsJSON(ctx) || IsPlain(ctx) {
		t.Fatalf("unexpected JSON mode flags")
	}

	ctx = WithMode(context.Background(), Mode{Plain: true})
	if IsJSON(ctx) || !IsPlain(ctx) {
		t.Fatalf("unexpected plain mode flags")
	}
}

func TestJSONTransformFromContext(t *testing.T) {
	if _, ok := JSONTransformFromContext(context.Background()); ok {
		t.Fatalf("expected missing transform in empty context")
	}

	ctx := WithJSONTransform(context.Background(), JSONTransform{
		ResultsOnly: true,
		Select:      []string{"id"},
	})
	transform, ok := JSONTransformFromContext(ctx)
	if !ok {
		t.Fatalf("expected transform in context")
	}
	if !transform.ResultsOnly || len(transform.Select) != 1 || transform.Select[0] != "id" {
		t.Fatalf("unexpected transform from context: %+v", transform)
	}
}

func TestWriteJSONErrorPaths(t *testing.T) {
	ctx := WithJSONTransform(context.Background(), JSONTransform{ResultsOnly: true})
	err := WriteJSON(ctx, io.Discard, map[string]any{"bad": func() {}})
	if err == nil || !strings.Contains(err.Error(), "transform json") {
		t.Fatalf("expected transform error, got %v", err)
	}

	err = WriteJSON(context.Background(), failingWriter{}, map[string]any{"ok": true})
	if err == nil || !strings.Contains(err.Error(), "encode json") {
		t.Fatalf("expected encode error, got %v", err)
	}
}

func TestTransformHelpers(t *testing.T) {
	if got := unwrapPrimary(map[string]any{"data": []any{1, 2}}); len(got.([]any)) != 2 {
		t.Fatalf("expected data unwrap, got %#v", got)
	}
	if got := unwrapPrimary(map[string]any{"results": []any{3, 4}}); len(got.([]any)) != 2 {
		t.Fatalf("expected results unwrap, got %#v", got)
	}
	if got := unwrapPrimary(map[string]any{"only": "value"}); got != "value" {
		t.Fatalf("expected single-key unwrap to value, got %#v", got)
	}
	multi := map[string]any{"a": 1, "b": 2}
	if got := unwrapPrimary(multi); got.(map[string]any)["a"] != 1 {
		t.Fatalf("expected multi-key map passthrough, got %#v", got)
	}

	items := []any{
		map[string]any{
			"name":  "Ada",
			"meta":  map[string]any{"active": true},
			"items": []any{"x", "y"},
		},
		map[string]any{
			"name":  "Bob",
			"meta":  map[string]any{"active": false},
			"items": []any{"z"},
		},
	}
	selected := selectFields(items, []string{"name", "meta.active", "items.0"})
	out, ok := selected.([]any)
	if !ok || len(out) != 2 {
		t.Fatalf("expected 2 selected items, got %#v", selected)
	}
	first := out[0].(map[string]any)
	if first["name"] != "Ada" || first["meta.active"] != true || first["items.0"] != "x" {
		t.Fatalf("unexpected selected item: %#v", first)
	}

	if got := selectFieldsFromItem("literal", []string{"name"}); got != "literal" {
		t.Fatalf("expected non-map passthrough, got %#v", got)
	}

	root := map[string]any{
		"workspace": map[string]any{
			"members": []any{
				map[string]any{"name": "Ada"},
			},
		},
	}
	if got, ok := getAtPath(root, "workspace.members.0.name"); !ok || got != "Ada" {
		t.Fatalf("expected nested path to resolve, got (%#v, %v)", got, ok)
	}
	if _, ok := getAtPath(root, "workspace.members.1.name"); ok {
		t.Fatalf("expected out-of-bounds array lookup to fail")
	}
	if _, ok := getAtPath(root, "workspace.members.bad.name"); ok {
		t.Fatalf("expected non-numeric array segment to fail")
	}
	if _, ok := getAtPath(root, "workspace..members"); ok {
		t.Fatalf("expected empty path segment to fail")
	}
	if _, ok := getAtPath(root, " "); ok {
		t.Fatalf("expected blank path to fail")
	}
	if _, ok := getAtPath(map[string]any{"a": "b"}, "a.c"); ok {
		t.Fatalf("expected scalar traversal to fail")
	}
}
