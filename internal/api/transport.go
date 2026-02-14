package api

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var sleep = time.Sleep

// RetryTransport retries 429 and 5xx responses.
type RetryTransport struct {
	Base       http.RoundTripper
	MaxRetries int
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	maxRetries := t.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		attemptReq, err := cloneRequest(req)
		if err != nil {
			return nil, fmt.Errorf("clone request: %w", err)
		}

		resp, err := base.RoundTrip(attemptReq)
		if err != nil {
			lastErr = err
			if attempt == maxRetries {
				return nil, err
			}
			wait := backoff(attempt)
			slog.Debug("request failed, retrying", "attempt", attempt+1, "wait", wait, "error", err)
			sleep(wait)
			continue
		}

		lastResp = resp
		if !shouldRetry(resp.StatusCode) || attempt == maxRetries {
			return resp, nil
		}

		wait := retryDelay(resp.Header.Get("Retry-After"), attempt)
		slog.Debug("retryable response", "status", resp.StatusCode, "attempt", attempt+1, "wait", wait)
		drainAndClose(resp.Body)
		sleep(wait)
	}

	if lastResp != nil {
		return lastResp, nil
	}
	return nil, lastErr
}

func shouldRetry(status int) bool {
	return status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	if d, ok := parseRetryAfter(retryAfter); ok {
		return d
	}
	return backoff(attempt)
}

func backoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	d := 250 * time.Millisecond
	for i := 0; i < attempt; i++ {
		d *= 2
		if d >= 5*time.Second {
			return 5 * time.Second
		}
	}
	return d
}

func parseRetryAfter(value string) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0, false
		}
		return time.Duration(seconds) * time.Second, true
	}
	if t, err := http.ParseTime(value); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d, true
		}
	}
	return 0, false
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	clone := req.Clone(req.Context())
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		clone.Body = body
	}
	return clone, nil
}

func drainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}
