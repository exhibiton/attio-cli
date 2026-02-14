package config

import (
	"path/filepath"
	"testing"
)

func setupConfigEnv(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("ATTIO_API_KEY", "")
	t.Setenv("ATTIO_BASE_URL", "")
	path := filepath.Join(tmp, "config.json")
	t.Setenv("ATTIO_CONFIG_PATH", path)
	return path
}

func TestLoadConfigMissingReturnsDefault(t *testing.T) {
	_ = setupConfigEnv(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultProfile != DefaultProfileName {
		t.Fatalf("expected default profile %q, got %q", DefaultProfileName, cfg.DefaultProfile)
	}
	if len(cfg.Profiles) != 0 {
		t.Fatalf("expected empty profiles, got %d", len(cfg.Profiles))
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	_ = setupConfigEnv(t)

	in := Config{
		DefaultProfile: "staging",
		Profiles: map[string]Profile{
			"staging": {APIKey: "cfg-key", BaseURL: "https://staging-api.attio.com"},
		},
	}
	if err := SaveConfig(in); err != nil {
		t.Fatalf("save config: %v", err)
	}

	out, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if out.DefaultProfile != "staging" {
		t.Fatalf("expected default profile staging, got %q", out.DefaultProfile)
	}
	if out.Profiles["staging"].APIKey != "cfg-key" {
		t.Fatalf("expected API key cfg-key, got %q", out.Profiles["staging"].APIKey)
	}
}

func TestResolveProfile(t *testing.T) {
	_ = setupConfigEnv(t)
	if got := ResolveProfile("custom"); got != "custom" {
		t.Fatalf("expected explicit profile custom, got %q", got)
	}

	if err := SaveConfig(Config{DefaultProfile: "from-config", Profiles: map[string]Profile{}}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if got := ResolveProfile(""); got != "from-config" {
		t.Fatalf("expected profile from config, got %q", got)
	}
}
