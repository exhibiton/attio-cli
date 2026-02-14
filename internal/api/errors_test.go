package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestParseAPIErrorSnakeCase(t *testing.T) {
	resp := &http.Response{
		StatusCode: 400,
		Header:     http.Header{"Retry-After": []string{"3"}},
		Body: io.NopCloser(strings.NewReader(`{
"status_code":400,
"type":"invalid_request_error",
"code":"validation_type",
"message":"invalid payload"
}`)),
	}

	err := parseAPIError(resp)
	attioErr, ok := err.(*AttioError)
	if !ok {
		t.Fatalf("expected *AttioError, got %T", err)
	}
	if attioErr.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", attioErr.StatusCode)
	}
	if attioErr.Code != "validation_type" {
		t.Fatalf("expected code validation_type, got %q", attioErr.Code)
	}
	if attioErr.Type != "invalid_request_error" {
		t.Fatalf("expected type invalid_request_error, got %q", attioErr.Type)
	}
	if attioErr.RetryAfter != "3" {
		t.Fatalf("expected retry-after 3, got %q", attioErr.RetryAfter)
	}
}

func TestParseAPIErrorCamelCase(t *testing.T) {
	resp := &http.Response{
		StatusCode: 502,
		Body: io.NopCloser(strings.NewReader(`{
"statusCode":429,
"type":"auth_error",
"code":"unauthorized",
"message":"bad token"
}`)),
	}

	err := parseAPIError(resp)
	attioErr, ok := err.(*AttioError)
	if !ok {
		t.Fatalf("expected *AttioError, got %T", err)
	}
	if attioErr.StatusCode != 429 {
		t.Fatalf("expected parsed status 429, got %d", attioErr.StatusCode)
	}
	if !IsAuthError(attioErr) {
		t.Fatalf("expected auth error")
	}
	if !IsRateLimited(attioErr) {
		t.Fatalf("expected rate-limited error")
	}
}

func TestParseAPIErrorFallbackMessage(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadGateway,
		Body:       io.NopCloser(strings.NewReader("upstream exploded")),
	}

	err := parseAPIError(resp)
	attioErr, ok := err.(*AttioError)
	if !ok {
		t.Fatalf("expected *AttioError, got %T", err)
	}
	if attioErr.Message != "upstream exploded" {
		t.Fatalf("expected fallback body message, got %q", attioErr.Message)
	}
}

func TestIsNotFound(t *testing.T) {
	err := &AttioError{StatusCode: 404, Code: "not_found", Message: "missing"}
	if !IsNotFound(err) {
		t.Fatalf("expected not-found helper to match")
	}
}

func TestAttioErrorErrorFormatting(t *testing.T) {
	var nilErr *AttioError
	if nilErr.Error() != "" {
		t.Fatalf("expected nil receiver to render empty string")
	}

	errWithCode := (&AttioError{StatusCode: 422, Code: "validation_error", Message: "bad data"}).Error()
	if !strings.Contains(errWithCode, "422 validation_error") || !strings.Contains(errWithCode, "bad data") {
		t.Fatalf("unexpected error string with code: %q", errWithCode)
	}

	errWithType := (&AttioError{StatusCode: 401, Type: "auth_error", Message: "unauthorized"}).Error()
	if !strings.Contains(errWithType, "401 auth_error") || !strings.Contains(errWithType, "unauthorized") {
		t.Fatalf("unexpected error string with type: %q", errWithType)
	}

	errFallback := (&AttioError{StatusCode: 500, Message: "upstream"}).Error()
	if !strings.Contains(errFallback, "500") || !strings.Contains(errFallback, "upstream") {
		t.Fatalf("unexpected fallback error string: %q", errFallback)
	}
}

func TestIntFromAny(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int
		ok   bool
	}{
		{name: "nil", in: nil, want: 0, ok: false},
		{name: "int", in: int(7), want: 7, ok: true},
		{name: "int32", in: int32(8), want: 8, ok: true},
		{name: "int64", in: int64(9), want: 9, ok: true},
		{name: "float64", in: float64(10), want: 10, ok: true},
		{name: "json-number", in: json.Number("11"), want: 11, ok: true},
		{name: "json-number-invalid", in: json.Number("11.2"), want: 0, ok: false},
		{name: "string", in: " 12 ", want: 12, ok: true},
		{name: "string-invalid", in: "abc", want: 0, ok: false},
		{name: "unsupported", in: struct{}{}, want: 0, ok: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, ok := intFromAny(tc.in)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("intFromAny(%#v) = (%d, %v), want (%d, %v)", tc.in, got, ok, tc.want, tc.ok)
			}
		})
	}
}
