package cmd

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/99designs/keyring"

	"github.com/failup-ventures/attio-cli/internal/config"
)

func stubInitPrompts(
	t *testing.T,
	line func(io.Writer, string, string) (string, error),
	secret func(io.Writer, string) (string, error),
	confirm func(io.Writer, string, bool) (bool, error),
) {
	t.Helper()

	origTTY := initIsTerminalFunc
	origLine := initPromptLineFunc
	origSecret := initPromptSecretFunc
	origConfirm := initPromptBoolFunc

	initIsTerminalFunc = func(fd int) bool { return true }
	if line != nil {
		initPromptLineFunc = line
	}
	if secret != nil {
		initPromptSecretFunc = secret
	}
	if confirm != nil {
		initPromptBoolFunc = confirm
	}

	t.Cleanup(func() {
		initIsTerminalFunc = origTTY
		initPromptLineFunc = origLine
		initPromptSecretFunc = origSecret
		initPromptBoolFunc = origConfirm
	})
}

func TestExecuteInitInteractiveJSONVerifiesAndSavesKey(t *testing.T) {
	setupCLIEnv(t)

	stubInitPrompts(
		t,
		func(_ io.Writer, question, defaultValue string) (string, error) {
			if strings.Contains(question, "Profile to configure") {
				return "onboarding", nil
			}
			return defaultValue, nil
		},
		func(_ io.Writer, _ string) (string, error) {
			return "init-key", nil
		},
		func(_ io.Writer, question string, defaultValue bool) (bool, error) {
			if strings.Contains(question, "Verify API key") {
				return true, nil
			}
			if strings.Contains(question, "Store API key") {
				return true, nil
			}
			return defaultValue, nil
		},
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/self" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer init-key" {
			t.Fatalf("expected auth header for init key, got %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"active":         true,
			"workspace_name": "Onboarding Workspace",
			"workspace_slug": "onboarding-workspace",
			"workspace_id":   "ws_123",
		})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "init"})
	if err != nil {
		t.Fatalf("init failed: %v stderr=%s", err, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal init output: %v output=%s", err, stdout)
	}
	if payload["initialized"] != true {
		t.Fatalf("expected initialized=true, got %#v", payload["initialized"])
	}
	if payload["profile"] != "onboarding" {
		t.Fatalf("unexpected profile: %#v", payload["profile"])
	}
	if payload["key_source"] != "prompt" {
		t.Fatalf("expected key_source prompt, got %#v", payload["key_source"])
	}
	if payload["saved_to_keyring"] != true {
		t.Fatalf("expected saved_to_keyring=true, got %#v", payload["saved_to_keyring"])
	}
	if payload["verified"] != true {
		t.Fatalf("expected verified=true, got %#v", payload["verified"])
	}
	if payload["workspace_name"] != "Onboarding Workspace" {
		t.Fatalf("unexpected workspace name: %#v", payload["workspace_name"])
	}

	stored, err := config.LoadAPIKey("onboarding")
	if err != nil {
		t.Fatalf("expected stored key for onboarding profile, got %v", err)
	}
	if stored != "init-key" {
		t.Fatalf("expected stored key init-key, got %q", stored)
	}
}

func TestExecuteInitInteractiveJSONSkipVerifyAndNoStore(t *testing.T) {
	setupCLIEnv(t)

	stubInitPrompts(
		t,
		func(_ io.Writer, _ string, defaultValue string) (string, error) {
			return defaultValue, nil
		},
		func(_ io.Writer, _ string) (string, error) {
			return "prompt-key", nil
		},
		func(_ io.Writer, question string, _ bool) (bool, error) {
			if strings.Contains(question, "Verify API key") {
				return false, nil
			}
			if strings.Contains(question, "Store API key") {
				return false, nil
			}
			return false, nil
		},
	)

	stdout, stderr, err := captureExecute(t, []string{"--json", "init"})
	if err != nil {
		t.Fatalf("init failed: %v stderr=%s", err, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal init output: %v output=%s", err, stdout)
	}
	if payload["saved_to_keyring"] != false {
		t.Fatalf("expected saved_to_keyring=false, got %#v", payload["saved_to_keyring"])
	}
	if payload["verified"] != false {
		t.Fatalf("expected verified=false, got %#v", payload["verified"])
	}

	_, err = config.LoadAPIKey("default")
	if !errors.Is(err, keyring.ErrKeyNotFound) {
		t.Fatalf("expected no stored default key, got %v", err)
	}
}

func TestExecuteInitNoInputRejected(t *testing.T) {
	setupCLIEnv(t)

	_, stderr, err := captureExecute(t, []string{"--json", "--no-input", "init"})
	if err == nil {
		t.Fatalf("expected init no-input error")
	}
	if ExitCode(err) != ExitCodeUsage {
		t.Fatalf("expected usage exit code %d, got %d", ExitCodeUsage, ExitCode(err))
	}

	var payload map[string]any
	if unmarshalErr := json.Unmarshal([]byte(stderr), &payload); unmarshalErr != nil {
		t.Fatalf("expected JSON error envelope, got stderr=%q err=%v", stderr, unmarshalErr)
	}
	errObj, _ := payload["error"].(map[string]any)
	if errObj["kind"] != "usage" {
		t.Fatalf("expected usage error kind, got %#v", errObj)
	}
}

func TestExecuteInitRequiresInteractiveTTY(t *testing.T) {
	setupCLIEnv(t)

	origTTY := initIsTerminalFunc
	initIsTerminalFunc = func(fd int) bool { return false }
	t.Cleanup(func() { initIsTerminalFunc = origTTY })

	_, stderr, err := captureExecute(t, []string{"--json", "init"})
	if err == nil {
		t.Fatalf("expected init tty error")
	}
	if ExitCode(err) != ExitCodeUsage {
		t.Fatalf("expected usage exit code %d, got %d", ExitCodeUsage, ExitCode(err))
	}
	if !strings.Contains(stderr, "interactive terminal") {
		t.Fatalf("expected interactive terminal guidance, got %q", stderr)
	}
}

func TestExecuteInitInteractiveEmptyKeyRejected(t *testing.T) {
	setupCLIEnv(t)

	stubInitPrompts(
		t,
		func(_ io.Writer, _ string, defaultValue string) (string, error) {
			return defaultValue, nil
		},
		func(_ io.Writer, _ string) (string, error) {
			return "", nil
		},
		func(_ io.Writer, _ string, defaultValue bool) (bool, error) {
			return defaultValue, nil
		},
	)

	_, stderr, err := captureExecute(t, []string{"--json", "init"})
	if err == nil {
		t.Fatalf("expected empty key error")
	}
	if ExitCode(err) != ExitCodeUsage {
		t.Fatalf("expected usage exit code %d, got %d", ExitCodeUsage, ExitCode(err))
	}
	if !strings.Contains(stderr, "API key cannot be empty") {
		t.Fatalf("expected empty key message, got %q", stderr)
	}
}
