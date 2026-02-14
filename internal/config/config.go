package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultProfileName = "default"
	DefaultBaseURL     = "https://api.attio.com"
)

type Profile struct {
	APIKey  string `json:"api_key,omitempty"`
	BaseURL string `json:"base_url,omitempty"`
}

type Config struct {
	DefaultProfile string             `json:"default_profile,omitempty"`
	Profiles       map[string]Profile `json:"profiles,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		DefaultProfile: DefaultProfileName,
		Profiles:       map[string]Profile{},
	}
}

func LoadConfig() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}

	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg := DefaultConfig()
		return cfg, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if cfg.DefaultProfile == "" {
		cfg.DefaultProfile = DefaultProfileName
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if cfg.DefaultProfile == "" {
		cfg.DefaultProfile = DefaultProfileName
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func ResolveProfile(profile string) string {
	profile = strings.TrimSpace(profile)
	if profile != "" {
		return profile
	}

	cfg, err := LoadConfig()
	if err == nil {
		if p := strings.TrimSpace(cfg.DefaultProfile); p != "" {
			return p
		}
	}
	return DefaultProfileName
}

func configPath() (string, error) {
	if override := strings.TrimSpace(os.Getenv("ATTIO_CONFIG_PATH")); override != "" {
		if strings.HasPrefix(override, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolve home dir: %w", err)
			}
			override = filepath.Join(home, strings.TrimPrefix(override, "~/"))
		}
		return override, nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(dir, "attio-cli", "config.json"), nil
}
