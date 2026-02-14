package errfmt

import (
	"errors"
	"strings"
	"testing"

	"github.com/failup-ventures/attio-cli/internal/api"
	"github.com/failup-ventures/attio-cli/internal/config"
)

func TestFormatAuthRequired(t *testing.T) {
	err := &config.AuthRequiredError{Message: "missing key"}
	if got := Format(err); got != "missing key" {
		t.Fatalf("expected auth required message, got %q", got)
	}
}

func TestFormatAttioErrorUnauthorized(t *testing.T) {
	err := &api.AttioError{StatusCode: 401, Code: "unauthorized", Message: "bad key"}
	got := Format(err)
	if !strings.Contains(got, "Attio API error (401)") {
		t.Fatalf("unexpected formatted message: %q", got)
	}
	if !strings.Contains(got, "attio auth login") {
		t.Fatalf("expected auth hint in message: %q", got)
	}
}

func TestUserFacingErrorWrap(t *testing.T) {
	cause := errors.New("boom")
	err := NewUserFacingError("friendly", cause)
	if err.Error() != "friendly" {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped cause")
	}
}
