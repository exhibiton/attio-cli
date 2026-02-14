package cmd

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/99designs/keyring"

	"github.com/failup-ventures/attio-cli/internal/config"
)

func TestExecuteInitJSONVerifiesAndSavesKey(t *testing.T) {
	setupCLIEnv(t)

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

	stdout, stderr, err := captureExecute(t, []string{"--json", "init", "--profile", "onboarding", "--api-key", "init-key"})
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
	if payload["key_source"] != "flag" {
		t.Fatalf("expected key_source flag, got %#v", payload["key_source"])
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

func TestExecuteInitJSONNoStoreAndSkipVerify(t *testing.T) {
	setupCLIEnv(t)

	stdout, stderr, err := captureExecute(t, []string{"--json", "--no-input", "init", "--api-key", "init-key", "--no-store", "--skip-verify"})
	if err != nil {
		t.Fatalf("init --no-store --skip-verify failed: %v stderr=%s", err, stderr)
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

func TestExecuteInitUsesEnvKeySource(t *testing.T) {
	setupCLIEnv(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/self" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer env-init-key" {
			t.Fatalf("expected auth header from env key, got %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"active": true, "workspace_name": "Env Workspace"})
	}))
	defer srv.Close()

	t.Setenv("ATTIO_API_KEY", "env-init-key")
	t.Setenv("ATTIO_BASE_URL", srv.URL)

	stdout, stderr, err := captureExecute(t, []string{"--json", "--no-input", "init", "--no-store"})
	if err != nil {
		t.Fatalf("init with env key failed: %v stderr=%s", err, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal init output: %v output=%s", err, stdout)
	}
	if payload["key_source"] != "env" {
		t.Fatalf("expected env key source, got %#v", payload["key_source"])
	}
	if payload["verified"] != true {
		t.Fatalf("expected verified=true, got %#v", payload["verified"])
	}
}

func TestExecuteInitNoInputMissingKey(t *testing.T) {
	setupCLIEnv(t)

	_, stderr, err := captureExecute(t, []string{"--json", "--no-input", "init"})
	if err == nil {
		t.Fatalf("expected init missing key error")
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
