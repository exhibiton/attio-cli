package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

type Mode struct {
	JSON  bool
	Plain bool
}

type ParseError struct{ msg string }

func (e *ParseError) Error() string { return e.msg }

func FromFlags(jsonOut bool, plainOut bool) (Mode, error) {
	if jsonOut && plainOut {
		return Mode{}, &ParseError{msg: "invalid output mode (cannot combine --json and --plain)"}
	}
	return Mode{JSON: jsonOut, Plain: plainOut}, nil
}

type modeKey struct{}

func WithMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, modeKey{}, mode)
}

func FromContext(ctx context.Context) Mode {
	if v := ctx.Value(modeKey{}); v != nil {
		if mode, ok := v.(Mode); ok {
			return mode
		}
	}
	return Mode{}
}

func IsJSON(ctx context.Context) bool  { return FromContext(ctx).JSON }
func IsPlain(ctx context.Context) bool { return FromContext(ctx).Plain }

func WriteJSON(ctx context.Context, w io.Writer, value any) error {
	if t, ok := JSONTransformFromContext(ctx); ok && (t.ResultsOnly || len(t.Select) > 0) {
		transformed, err := applyJSONTransform(value, t)
		if err != nil {
			return fmt.Errorf("transform json: %w", err)
		}
		value = transformed
	}

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
