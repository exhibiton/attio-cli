package errfmt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/99designs/keyring"
	"github.com/alecthomas/kong"

	"github.com/failup-ventures/attio-cli/internal/api"
	"github.com/failup-ventures/attio-cli/internal/config"
)

func Format(err error) string {
	if err == nil {
		return ""
	}

	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		return formatParseError(parseErr)
	}

	var authErr *config.AuthRequiredError
	if errors.As(err, &authErr) {
		return authErr.Error()
	}

	if errors.Is(err, keyring.ErrKeyNotFound) {
		return "No API key in keyring for this profile. Run: attio auth login --api-key <key>"
	}

	var attioErr *api.AttioError
	if errors.As(err, &attioErr) {
		msg := fmt.Sprintf("Attio API error (%d)", attioErr.StatusCode)
		if attioErr.Code != "" {
			msg += " " + attioErr.Code
		}
		if attioErr.Message != "" {
			msg += ": " + attioErr.Message
		}
		if attioErr.StatusCode == 429 && attioErr.RetryAfter != "" {
			msg += fmt.Sprintf(" (retry-after: %s)", attioErr.RetryAfter)
		}
		if attioErr.StatusCode == 401 || attioErr.StatusCode == 403 {
			msg += "\nCheck ATTIO_API_KEY or run: attio auth login --api-key <key>"
		}
		return msg
	}

	return strings.TrimSpace(err.Error())
}

// UserFacingError preserves a friendly message while keeping a cause.
type UserFacingError struct {
	Message string
	Cause   error
}

func (e *UserFacingError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *UserFacingError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func NewUserFacingError(message string, cause error) error {
	return &UserFacingError{Message: message, Cause: cause}
}

func formatParseError(err *kong.ParseError) string {
	msg := err.Error()
	if strings.Contains(msg, "did you mean") {
		return msg
	}
	if strings.HasPrefix(msg, "unknown flag") {
		return msg + "\nRun with --help to see available flags"
	}
	if strings.Contains(msg, "missing") || strings.Contains(msg, "required") {
		return msg + "\nRun with --help to see usage"
	}
	return msg
}
