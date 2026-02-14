package api

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestRetryTransportRetries429ThenSuccess(t *testing.T) {
	origSleep := sleep
	sleep = func(time.Duration) {}
	t.Cleanup(func() { sleep = origSleep })

	attempts := 0
	rt := &RetryTransport{
		Base: roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			attempts++
			if attempts == 1 {
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Header:     http.Header{"Retry-After": []string{"1"}},
					Body:       io.NopCloser(strings.NewReader("rate limited")),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
			}, nil
		}),
		MaxRetries: 1,
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRetryTransportNoRetryFor400(t *testing.T) {
	attempts := 0
	rt := &RetryTransport{
		Base: roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			attempts++
			return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("bad"))}, nil
		}),
		MaxRetries: 3,
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestParseRetryAfter(t *testing.T) {
	if d, ok := parseRetryAfter("2"); !ok || d != 2*time.Second {
		t.Fatalf("expected 2s, got %v ok=%v", d, ok)
	}
	if _, ok := parseRetryAfter("0"); ok {
		t.Fatalf("expected invalid retry-after for 0")
	}
	if _, ok := parseRetryAfter("not-a-duration"); ok {
		t.Fatalf("expected invalid retry-after")
	}
}

func TestBackoffCaps(t *testing.T) {
	if got := backoff(0); got != 250*time.Millisecond {
		t.Fatalf("expected 250ms, got %v", got)
	}
	if got := backoff(10); got != 5*time.Second {
		t.Fatalf("expected capped 5s, got %v", got)
	}
}

func TestRetryTransportRetriesErrorThenSuccess(t *testing.T) {
	origSleep := sleep
	sleep = func(time.Duration) {}
	t.Cleanup(func() { sleep = origSleep })

	attempts := 0
	rt := &RetryTransport{
		Base: roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			attempts++
			if attempts == 1 {
				return nil, errors.New("temporary network failure")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
			}, nil
		}),
		MaxRetries: 2,
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestRetryTransportCloneRequestError(t *testing.T) {
	rt := &RetryTransport{
		Base: roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			t.Fatalf("base transport should not be called when cloning fails")
			return nil, nil
		}),
		MaxRetries: 1,
	}

	req, _ := http.NewRequest(http.MethodPost, "https://example.com", strings.NewReader("x"))
	req.GetBody = func() (io.ReadCloser, error) {
		return nil, errors.New("getbody failed")
	}

	_, err := rt.RoundTrip(req)
	if err == nil || !strings.Contains(err.Error(), "clone request") {
		t.Fatalf("expected clone request error, got %v", err)
	}
}

func TestRetryTransportNegativeRetriesNoRetry(t *testing.T) {
	attempts := 0
	rt := &RetryTransport{
		Base: roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			attempts++
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("server error")),
			}, nil
		}),
		MaxRetries: -1,
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if attempts != 1 {
		t.Fatalf("expected exactly one attempt with negative retries, got %d", attempts)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 response, got %d", resp.StatusCode)
	}
}

func TestRetryDelayFallbackAndParseRetryAfterDate(t *testing.T) {
	if got := retryDelay("invalid", 2); got != backoff(2) {
		t.Fatalf("expected backoff fallback, got %v", got)
	}

	future := time.Now().Add(2 * time.Second).UTC().Format(http.TimeFormat)
	if d, ok := parseRetryAfter(future); !ok || d <= 0 {
		t.Fatalf("expected positive retry-after duration from HTTP date, got %v ok=%v", d, ok)
	}

	past := time.Now().Add(-2 * time.Second).UTC().Format(http.TimeFormat)
	if _, ok := parseRetryAfter(past); ok {
		t.Fatalf("expected past retry-after date to be rejected")
	}
}

type trackedReadCloser struct {
	closed bool
	body   io.Reader
}

func (t *trackedReadCloser) Read(p []byte) (int, error) {
	if t.body == nil {
		return 0, io.EOF
	}
	return t.body.Read(p)
}

func (t *trackedReadCloser) Close() error {
	t.closed = true
	return nil
}

func TestDrainAndClose(t *testing.T) {
	drainAndClose(nil)

	rc := &trackedReadCloser{body: strings.NewReader("payload")}
	drainAndClose(rc)
	if !rc.closed {
		t.Fatalf("expected read closer to be closed")
	}
}
