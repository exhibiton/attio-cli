package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type JSONTransform struct {
	ResultsOnly bool
	Select      []string
}

type jsonTransformKey struct{}

func WithJSONTransform(ctx context.Context, transform JSONTransform) context.Context {
	return context.WithValue(ctx, jsonTransformKey{}, transform)
}

func JSONTransformFromContext(ctx context.Context) (JSONTransform, bool) {
	v := ctx.Value(jsonTransformKey{})
	if v == nil {
		return JSONTransform{}, false
	}
	transform, ok := v.(JSONTransform)
	return transform, ok
}

func applyJSONTransform(value any, transform JSONTransform) (any, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	var anyV any
	if err := json.Unmarshal(b, &anyV); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	if transform.ResultsOnly {
		anyV = unwrapPrimary(anyV)
	}
	if len(transform.Select) > 0 {
		anyV = selectFields(anyV, transform.Select)
	}
	return anyV, nil
}

func unwrapPrimary(v any) any {
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}
	if data, ok := m["data"]; ok {
		return data
	}
	if results, ok := m["results"]; ok {
		return results
	}
	if len(m) == 1 {
		for _, val := range m {
			return val
		}
	}
	return v
}

func selectFields(v any, fields []string) any {
	switch vv := v.(type) {
	case []any:
		out := make([]any, 0, len(vv))
		for _, item := range vv {
			out = append(out, selectFieldsFromItem(item, fields))
		}
		return out
	default:
		return selectFieldsFromItem(v, fields)
	}
}

func selectFieldsFromItem(v any, fields []string) any {
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}

	out := make(map[string]any, len(fields))
	for _, field := range fields {
		if value, ok := getAtPath(m, field); ok {
			out[field] = value
		}
	}
	return out
}

func getAtPath(v any, path string) (any, bool) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, false
	}

	segments := strings.Split(path, ".")
	current := v
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			return nil, false
		}
		switch c := current.(type) {
		case map[string]any:
			next, ok := c[segment]
			if !ok {
				return nil, false
			}
			current = next
		case []any:
			i, err := strconv.Atoi(segment)
			if err != nil || i < 0 || i >= len(c) {
				return nil, false
			}
			current = c[i]
		default:
			return nil, false
		}
	}

	return current, true
}
