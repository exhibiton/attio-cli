package config

import (
	"errors"
	"testing"
)

func TestResolveAPIKeyFromEnvTakesPrecedence(t *testing.T) {
	_ = setupConfigEnv(t)

	if err := StoreAPIKey("default", "keyring-key"); err != nil {
		t.Fatalf("store keyring key: %v", err)
	}
	if err := SaveConfig(Config{DefaultProfile: "default", Profiles: map[string]Profile{"default": {APIKey: "config-key"}}}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	t.Setenv("ATTIO_API_KEY", "env-key")

	key, source, err := ResolveAPIKeyWithSource("default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "env-key" || source != AuthSourceEnv {
		t.Fatalf("expected env source, got key=%q source=%q", key, source)
	}
}

func TestResolveAPIKeyFromKeyringFallback(t *testing.T) {
	_ = setupConfigEnv(t)

	if err := StoreAPIKey("default", "keyring-key"); err != nil {
		t.Fatalf("store keyring key: %v", err)
	}

	key, source, err := ResolveAPIKeyWithSource("default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "keyring-key" || source != AuthSourceKeyring {
		t.Fatalf("expected keyring source, got key=%q source=%q", key, source)
	}
}

func TestResolveAPIKeyFromConfigFallback(t *testing.T) {
	_ = setupConfigEnv(t)

	if err := SaveConfig(Config{DefaultProfile: "default", Profiles: map[string]Profile{"default": {APIKey: "config-key"}}}); err != nil {
		t.Fatalf("save config: %v", err)
	}

	key, source, err := ResolveAPIKeyWithSource("default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "config-key" || source != AuthSourceConfig {
		t.Fatalf("expected config source, got key=%q source=%q", key, source)
	}
}

func TestResolveAPIKeyMissingReturnsAuthRequired(t *testing.T) {
	_ = setupConfigEnv(t)

	_, _, err := ResolveAPIKeyWithSource("default")
	if err == nil {
		t.Fatalf("expected error")
	}
	var authErr *AuthRequiredError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthRequiredError, got %T", err)
	}
}

func TestResolveBaseURLPrecedence(t *testing.T) {
	_ = setupConfigEnv(t)

	if err := SaveConfig(Config{DefaultProfile: "default", Profiles: map[string]Profile{"default": {BaseURL: "https://cfg.attio.test"}}}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if got := ResolveBaseURL("default"); got != "https://cfg.attio.test" {
		t.Fatalf("expected config base URL, got %q", got)
	}

	t.Setenv("ATTIO_BASE_URL", "https://env.attio.test/")
	if got := ResolveBaseURL("default"); got != "https://env.attio.test" {
		t.Fatalf("expected env base URL override, got %q", got)
	}
}

func TestAuthStatusResolved(t *testing.T) {
	_ = setupConfigEnv(t)

	if err := StoreAPIKey("default", "super-secret-key"); err != nil {
		t.Fatalf("store keyring key: %v", err)
	}

	status := AuthStatus("default")
	if !status.Resolved {
		t.Fatalf("expected resolved=true")
	}
	if status.ResolvedSource != AuthSourceKeyring {
		t.Fatalf("expected keyring source, got %q", status.ResolvedSource)
	}
	if status.MaskedKey == "" || status.MaskedKey == "super-secret-key" {
		t.Fatalf("expected masked key, got %q", status.MaskedKey)
	}
}
