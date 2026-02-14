package ui

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
)

func TestParseErrorError(t *testing.T) {
	err := &ParseError{msg: "bad color"}
	if err.Error() != "bad color" {
		t.Fatalf("unexpected parse error message: %q", err.Error())
	}
}

func TestNewInvalidColor(t *testing.T) {
	_, err := New(Options{Color: "rainbow"})
	if err == nil {
		t.Fatalf("expected parse error")
	}
	if _, ok := err.(*ParseError); !ok {
		t.Fatalf("expected ParseError, got %T", err)
	}
}

func TestNewColorNeverAndPrinterOutput(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	u, err := New(Options{Stdout: &out, Stderr: &errOut, Color: "never"})
	if err != nil {
		t.Fatalf("new ui: %v", err)
	}
	if u.Out().ColorEnabled() {
		t.Fatalf("expected color disabled for --color=never")
	}

	u.Out().Print("a")
	u.Out().Println("b")
	u.Out().Printf("%s", "c")
	u.Out().Successf("%s", "ok")
	u.Err().Error("bad")
	u.Err().Errorf("bad %d", 2)

	if got := out.String(); got != "ab\nc\nok\n" {
		t.Fatalf("unexpected stdout content: %q", got)
	}
	if got := errOut.String(); got != "bad\nbad 2\n" {
		t.Fatalf("unexpected stderr content: %q", got)
	}
}

func TestNoColorEnvForcesAscii(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	u, err := New(Options{Color: "always"})
	if err != nil {
		t.Fatalf("new ui: %v", err)
	}
	if u.Out().ColorEnabled() {
		t.Fatalf("expected NO_COLOR to disable colors")
	}
}

func TestWritersFallbackAndForwarding(t *testing.T) {
	var nilUI *UI
	if got := nilUI.OutWriter(); got != os.Stdout {
		t.Fatalf("expected nil ui out writer to fallback to os.Stdout")
	}
	if got := nilUI.ErrWriter(); got != os.Stderr {
		t.Fatalf("expected nil ui err writer to fallback to os.Stderr")
	}

	u := &UI{}
	if got := u.OutWriter(); got != os.Stdout {
		t.Fatalf("expected empty ui out writer to fallback to os.Stdout")
	}
	if got := u.ErrWriter(); got != os.Stderr {
		t.Fatalf("expected empty ui err writer to fallback to os.Stderr")
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	u2, err := New(Options{Stdout: &out, Stderr: &errOut, Color: "never"})
	if err != nil {
		t.Fatalf("new ui: %v", err)
	}

	_, _ = fmt.Fprint(u2.OutWriter(), "x")
	_, _ = fmt.Fprint(u2.ErrWriter(), "y")
	if out.String() != "x" {
		t.Fatalf("unexpected forwarded stdout: %q", out.String())
	}
	if errOut.String() != "y" {
		t.Fatalf("unexpected forwarded stderr: %q", errOut.String())
	}
}

func TestContextHelpers(t *testing.T) {
	u, err := New(Options{Color: "never"})
	if err != nil {
		t.Fatalf("new ui: %v", err)
	}

	ctx := WithUI(context.Background(), u)
	if got := FromContext(ctx); got != u {
		t.Fatalf("expected ui instance roundtrip")
	}
	if got := FromContext(context.Background()); got != nil {
		t.Fatalf("expected nil ui from empty context")
	}

	ctx = context.WithValue(context.Background(), ctxKey{}, "not-ui")
	if got := FromContext(ctx); got != nil {
		t.Fatalf("expected nil for invalid context value")
	}
}
