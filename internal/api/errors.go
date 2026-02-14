package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

// AttioError is a structured Attio API error.
type AttioError struct {
	StatusCode int
	Type       string
	Code       string
	Message    string
	RetryAfter string
}

func (e *AttioError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code != "" {
		return fmt.Sprintf("attio api error (%d %s): %s", e.StatusCode, e.Code, e.Message)
	}
	if e.Type != "" {
		return fmt.Sprintf("attio api error (%d %s): %s", e.StatusCode, e.Type, e.Message)
	}
	return fmt.Sprintf("attio api error (%d): %s", e.StatusCode, e.Message)
}

func parseAPIError(resp *http.Response) error {
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("read error response: %w", readErr)
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		slog.Debug("failed to parse API error response as JSON", "error", err)
	}

	ae := &AttioError{
		StatusCode: resp.StatusCode,
		RetryAfter: strings.TrimSpace(resp.Header.Get("Retry-After")),
	}

	if status, ok := intFromAny(raw["status_code"]); ok {
		ae.StatusCode = status
	} else if status, ok := intFromAny(raw["statusCode"]); ok {
		ae.StatusCode = status
	}
	if s, ok := stringFromAny(raw["type"]); ok {
		ae.Type = s
	}
	if s, ok := stringFromAny(raw["code"]); ok {
		ae.Code = s
	}
	if s, ok := stringFromAny(raw["message"]); ok {
		ae.Message = s
	}

	if ae.Message == "" {
		trimmed := strings.TrimSpace(string(body))
		if trimmed != "" {
			ae.Message = trimmed
		} else {
			ae.Message = http.StatusText(resp.StatusCode)
		}
	}

	return ae
}

func IsNotFound(err error) bool {
	var ae *AttioError
	return errors.As(err, &ae) && (ae.Code == "not_found" || ae.StatusCode == http.StatusNotFound)
}

func IsAuthError(err error) bool {
	var ae *AttioError
	if !errors.As(err, &ae) {
		return false
	}
	return ae.Type == "auth_error" || ae.StatusCode == http.StatusUnauthorized || ae.StatusCode == http.StatusForbidden
}

func IsRateLimited(err error) bool {
	var ae *AttioError
	return errors.As(err, &ae) && ae.StatusCode == http.StatusTooManyRequests
}

func intFromAny(v any) (int, bool) {
	switch x := v.(type) {
	case nil:
		return 0, false
	case int:
		return x, true
	case int32:
		return int(x), true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	case json.Number:
		i, err := x.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(x))
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

func stringFromAny(v any) (string, bool) {
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	return s, true
}
