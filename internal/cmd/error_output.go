package cmd

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/99designs/keyring"

	"github.com/failup-ventures/attio-cli/internal/api"
	"github.com/failup-ventures/attio-cli/internal/config"
	"github.com/failup-ventures/attio-cli/internal/errfmt"
)

func writeErrorJSON(w io.Writer, err error) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if encodeErr := enc.Encode(map[string]any{"error": errorObject(err)}); encodeErr != nil {
		_ = enc.Encode(map[string]any{
			"error": map[string]any{
				"message":   strings.TrimSpace(err.Error()),
				"exit_code": ExitCode(err),
			},
		})
	}
}

func errorObject(err error) map[string]any {
	msg := strings.TrimSpace(errfmt.Format(err))
	if msg == "" {
		msg = strings.TrimSpace(err.Error())
	}
	payload := map[string]any{
		"message":   msg,
		"exit_code": ExitCode(err),
	}

	var usageErr *UsageError
	if errors.As(err, &usageErr) {
		payload["kind"] = "usage"
	}

	var authErr *config.AuthRequiredError
	if errors.As(err, &authErr) {
		payload["kind"] = "auth_required"
		payload["code"] = "auth_required"
	}

	if errors.Is(err, keyring.ErrKeyNotFound) {
		payload["kind"] = "keyring"
		payload["code"] = "key_not_found"
	}

	var attioErr *api.AttioError
	if errors.As(err, &attioErr) {
		payload["kind"] = "api"
		if attioErr.StatusCode != 0 {
			payload["status_code"] = attioErr.StatusCode
		}
		if attioErr.Type != "" {
			payload["type"] = attioErr.Type
		}
		if attioErr.Code != "" {
			payload["code"] = attioErr.Code
		}
		if attioErr.Message != "" {
			payload["message"] = attioErr.Message
		}
		if attioErr.RetryAfter != "" {
			payload["retry_after"] = attioErr.RetryAfter
		}
	}

	return payload
}

func jsonRequestedFromArgs(args []string) bool {
	jsonRequested := false
	plainRequested := false

	for _, arg := range args {
		switch {
		case arg == "--json" || arg == "-j":
			jsonRequested = true
		case arg == "--plain" || arg == "-p":
			plainRequested = true
		case strings.HasPrefix(arg, "--json="):
			v := strings.TrimSpace(strings.TrimPrefix(arg, "--json="))
			if b, ok := parseBoolArg(v); ok {
				jsonRequested = b
			}
		case strings.HasPrefix(arg, "--plain="):
			v := strings.TrimSpace(strings.TrimPrefix(arg, "--plain="))
			if b, ok := parseBoolArg(v); ok {
				plainRequested = b
			}
		}
	}

	return jsonRequested && !plainRequested
}

func parseBoolArg(v string) (bool, bool) {
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return false, false
	}
	return parsed, true
}
