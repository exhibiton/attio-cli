package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupCLIEnv(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("ATTIO_CONFIG_PATH", filepath.Join(tmp, "config.json"))
	t.Setenv("ATTIO_API_KEY", "")
	t.Setenv("ATTIO_BASE_URL", "")
}

func captureExecute(t *testing.T, args []string) (string, string, error) {
	t.Helper()

	origOut := os.Stdout
	origErr := os.Stderr

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}

	os.Stdout = outW
	os.Stderr = errW

	execErr := Execute(args)

	_ = outW.Close()
	_ = errW.Close()
	os.Stdout = origOut
	os.Stderr = origErr

	outBytes, _ := io.ReadAll(outR)
	errBytes, _ := io.ReadAll(errR)
	_ = outR.Close()
	_ = errR.Close()

	return string(outBytes), string(errBytes), execErr
}

func TestExecuteVersionJSON(t *testing.T) {
	setupCLIEnv(t)

	stdout, stderr, err := captureExecute(t, []string{"--json", "version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal stdout: %v\noutput=%s", err, stdout)
	}
	if payload["version"] == "" {
		t.Fatalf("expected version key in output: %#v", payload)
	}
}

func TestExecuteSelfJSON(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/self" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer env-key" {
			t.Fatalf("unexpected auth header: %q", got)
		}
		_, _ = w.Write([]byte(`{"active":true,"workspace_name":"Test WS","workspace_slug":"test-ws"}`))
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "self"})
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr=%s", err, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, `"workspace_name": "Test WS"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestExecuteAuthStatusJSON(t *testing.T) {
	setupCLIEnv(t)
	t.Setenv("ATTIO_API_KEY", "env-key")

	stdout, stderr, err := captureExecute(t, []string{"--json", "auth", "status"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal stdout: %v\noutput=%s", err, stdout)
	}
	if payload["resolved"] != true {
		t.Fatalf("expected resolved=true, got %#v", payload)
	}
	if payload["resolved_source"] != "env" {
		t.Fatalf("expected resolved_source env, got %#v", payload)
	}
}

func TestExecuteSelfAuthMissing(t *testing.T) {
	setupCLIEnv(t)

	_, stderr, err := captureExecute(t, []string{"--json", "self"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if ExitCode(err) != ExitCodeAuth {
		t.Fatalf("expected auth exit code %d, got %d", ExitCodeAuth, ExitCode(err))
	}
	if !bytes.Contains([]byte(stderr), []byte("No API key found")) {
		t.Fatalf("expected auth error message, got %q", stderr)
	}
}
