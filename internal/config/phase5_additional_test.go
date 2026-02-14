package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/99designs/keyring"
)

func TestAuthRequiredErrorError(t *testing.T) {
	var nilErr *AuthRequiredError
	if nilErr.Error() != "" {
		t.Fatalf("expected nil error string to be empty")
	}

	err := &AuthRequiredError{Message: "missing key"}
	if err.Error() != "missing key" {
		t.Fatalf("unexpected auth error message: %q", err.Error())
	}
}

func TestResolveAPIKeyWrapper(t *testing.T) {
	_ = setupConfigEnv(t)
	t.Setenv("ATTIO_API_KEY", "env-key")

	key, err := ResolveAPIKey("default")
	if err != nil {
		t.Fatalf("resolve api key: %v", err)
	}
	if key != "env-key" {
		t.Fatalf("expected env-key, got %q", key)
	}
}

func TestSaveConfigNormalizesDefaults(t *testing.T) {
	path := setupConfigEnv(t)

	if err := SaveConfig(Config{}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.DefaultProfile != DefaultProfileName {
		t.Fatalf("expected default profile %q, got %q", DefaultProfileName, cfg.DefaultProfile)
	}
	if cfg.Profiles == nil {
		t.Fatalf("expected profiles map to be initialized")
	}
}

func TestLoadConfigParseError(t *testing.T) {
	path := setupConfigEnv(t)
	if err := os.WriteFile(path, []byte("{bad json"), 0o600); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	_, err := LoadConfig()
	if err == nil {
		t.Fatalf("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse config") {
		t.Fatalf("expected parse config error, got %v", err)
	}
}

func TestConfigPathExpandTildeOverride(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("ATTIO_CONFIG_PATH", "~/nested/config.json")

	path, err := configPath()
	if err != nil {
		t.Fatalf("configPath: %v", err)
	}
	want := filepath.Join(home, "nested", "config.json")
	if path != want {
		t.Fatalf("expected %q, got %q", want, path)
	}
}

func TestRemoveAPIKey(t *testing.T) {
	_ = setupConfigEnv(t)

	if err := StoreAPIKey("default", "to-remove"); err != nil {
		t.Fatalf("store keyring key: %v", err)
	}
	if err := RemoveAPIKey("default"); err != nil {
		t.Fatalf("remove keyring key: %v", err)
	}

	_, err := LoadAPIKey("default")
	if !errors.Is(err, keyring.ErrKeyNotFound) {
		t.Fatalf("expected keyring.ErrKeyNotFound after removal, got %v", err)
	}
}

func TestConfigPathPlainOverrideAndDefaultBranch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	override := filepath.Join(home, "custom", "config.json")
	t.Setenv("ATTIO_CONFIG_PATH", override)

	path, err := configPath()
	if err != nil {
		t.Fatalf("configPath override: %v", err)
	}
	if path != override {
		t.Fatalf("expected override path %q, got %q", override, path)
	}

	t.Setenv("ATTIO_CONFIG_PATH", "")
	path, err = configPath()
	if err != nil {
		t.Fatalf("configPath default branch: %v", err)
	}
	if !strings.HasSuffix(path, filepath.Join("attio-cli", "config.json")) {
		t.Fatalf("expected default config suffix, got %q", path)
	}

	// On macOS and Linux, UserConfigDir should include HOME for test-local determinism.
	if runtime.GOOS != "windows" && !strings.Contains(path, home) {
		t.Fatalf("expected default config path to include test HOME %q, got %q", home, path)
	}
}

func TestLoadConfigBackfillsDefaultsFromSparseFile(t *testing.T) {
	path := setupConfigEnv(t)
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatalf("write sparse config: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("load sparse config: %v", err)
	}
	if cfg.DefaultProfile != DefaultProfileName {
		t.Fatalf("expected default profile %q, got %q", DefaultProfileName, cfg.DefaultProfile)
	}
	if cfg.Profiles == nil {
		t.Fatalf("expected profiles map to be initialized")
	}
}

func TestResolveProfileFallsBackWhenConfigUnreadable(t *testing.T) {
	path := setupConfigEnv(t)
	if err := os.WriteFile(path, []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	if got := ResolveProfile(""); got != DefaultProfileName {
		t.Fatalf("expected fallback profile %q, got %q", DefaultProfileName, got)
	}
}

func TestMaskKeyEdgeCases(t *testing.T) {
	if got := maskKey(""); got != "" {
		t.Fatalf("expected empty masked key, got %q", got)
	}
	if got := maskKey("abc123"); got != "******" {
		t.Fatalf("expected full masking for short key, got %q", got)
	}
	if got := maskKey("super-secret-key"); !strings.HasSuffix(got, "-key") {
		t.Fatalf("expected long key to preserve last 4 chars, got %q", got)
	}
}
