package errfmt

import (
	"errors"
	"strings"
	"testing"

	"github.com/99designs/keyring"
	"github.com/alecthomas/kong"

	"github.com/failup-ventures/attio-cli/internal/api"
)

func mustParseErr(t *testing.T, model any, args ...string) error {
	t.Helper()
	parser, err := kong.New(model)
	if err != nil {
		t.Fatalf("kong.New: %v", err)
	}
	_, err = parser.Parse(args)
	if err == nil {
		t.Fatalf("expected parse error")
	}
	return err
}

func TestFormatNilAndGeneric(t *testing.T) {
	if got := Format(nil); got != "" {
		t.Fatalf("expected empty string for nil error, got %q", got)
	}

	err := errors.New("  spaced message  ")
	if got := Format(err); got != "spaced message" {
		t.Fatalf("expected trimmed generic message, got %q", got)
	}
}

func TestFormatKeyringNotFound(t *testing.T) {
	got := Format(keyring.ErrKeyNotFound)
	if got != "No API key in keyring for this profile. Run: attio auth login --api-key <key>" {
		t.Fatalf("unexpected keyring message: %q", got)
	}
}

func TestFormatParseErrorVariants(t *testing.T) {
	t.Run("unknown-flag", func(t *testing.T) {
		var cli struct {
			Verbose bool `name:"verbose"`
		}
		err := mustParseErr(t, &cli, "--bogus")
		got := Format(err)
		if !strings.Contains(got, "unknown flag") {
			t.Fatalf("expected unknown flag message, got %q", got)
		}
		if !strings.Contains(got, "Run with --help to see available flags") {
			t.Fatalf("expected available-flags help, got %q", got)
		}
	})

	t.Run("missing-required", func(t *testing.T) {
		var cli struct {
			Token string `name:"token" required:""`
		}
		err := mustParseErr(t, &cli)
		got := Format(err)
		if !strings.Contains(got, "missing flags") {
			t.Fatalf("expected missing-flags message, got %q", got)
		}
		if !strings.Contains(got, "Run with --help to see usage") {
			t.Fatalf("expected usage help, got %q", got)
		}
	})

	t.Run("did-you-mean", func(t *testing.T) {
		var cli struct {
			Hello struct{} `cmd:""`
		}
		err := mustParseErr(t, &cli, "helo")
		got := Format(err)
		if !strings.Contains(got, "did you mean") {
			t.Fatalf("expected did-you-mean parse message, got %q", got)
		}
		if got != err.Error() {
			t.Fatalf("did-you-mean parse messages should not be altered")
		}
	})

	t.Run("default", func(t *testing.T) {
		var cli struct {
			Hello struct{} `cmd:""`
		}
		err := mustParseErr(t, &cli)
		got := Format(err)
		if got != err.Error() {
			t.Fatalf("expected default parse message passthrough, got %q (want %q)", got, err.Error())
		}
	})
}

func TestFormatAttioRateLimited(t *testing.T) {
	err := &api.AttioError{
		StatusCode: 429,
		Code:       "rate_limited",
		Message:    "too many requests",
		RetryAfter: "2",
	}
	got := Format(err)
	if !strings.Contains(got, "Attio API error (429)") || !strings.Contains(got, "retry-after: 2") {
		t.Fatalf("unexpected rate-limited format: %q", got)
	}
}

func TestUserFacingErrorNilMethods(t *testing.T) {
	var err *UserFacingError
	if err.Error() != "" {
		t.Fatalf("expected nil error string to be empty")
	}
	if err.Unwrap() != nil {
		t.Fatalf("expected nil unwrap for nil receiver")
	}
}
