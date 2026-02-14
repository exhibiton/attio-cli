package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/99designs/keyring"
)

type AuthSource string

const (
	AuthSourceEnv     AuthSource = "env"
	AuthSourceKeyring AuthSource = "keyring"
	AuthSourceConfig  AuthSource = "config"
)

type AuthRequiredError struct {
	Message string
}

func (e *AuthRequiredError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

type Status struct {
	Profile        string     `json:"profile"`
	BaseURL        string     `json:"base_url"`
	HasEnv         bool       `json:"has_env"`
	HasKeyring     bool       `json:"has_keyring"`
	HasConfig      bool       `json:"has_config"`
	Resolved       bool       `json:"resolved"`
	ResolvedSource AuthSource `json:"resolved_source,omitempty"`
	MaskedKey      string     `json:"masked_key,omitempty"`
}

func ResolveAPIKey(profile string) (string, error) {
	key, _, err := ResolveAPIKeyWithSource(profile)
	return key, err
}

func ResolveAPIKeyWithSource(profile string) (string, AuthSource, error) {
	profile = ResolveProfile(profile)

	if key := strings.TrimSpace(os.Getenv("ATTIO_API_KEY")); key != "" {
		return key, AuthSourceEnv, nil
	}

	if key, err := LoadAPIKey(profile); err == nil && key != "" {
		return key, AuthSourceKeyring, nil
	} else if err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
		// Continue to config file fallback instead of hard-failing if keyring is unavailable.
	}

	cfg, err := LoadConfig()
	if err == nil {
		if p, ok := cfg.Profiles[profile]; ok {
			if key := strings.TrimSpace(p.APIKey); key != "" {
				return key, AuthSourceConfig, nil
			}
		}
	}

	return "", "", &AuthRequiredError{
		Message: fmt.Sprintf("No API key found for profile %q. Set ATTIO_API_KEY or run: attio auth login --api-key <key>", profile),
	}
}

func ResolveBaseURL(profile string) string {
	if v := strings.TrimSpace(os.Getenv("ATTIO_BASE_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}

	profile = ResolveProfile(profile)
	cfg, err := LoadConfig()
	if err == nil {
		if p, ok := cfg.Profiles[profile]; ok {
			if v := strings.TrimSpace(p.BaseURL); v != "" {
				return strings.TrimRight(v, "/")
			}
		}
	}

	return DefaultBaseURL
}

func AuthStatus(profile string) Status {
	profile = ResolveProfile(profile)
	status := Status{
		Profile: profile,
		BaseURL: ResolveBaseURL(profile),
	}

	if key := strings.TrimSpace(os.Getenv("ATTIO_API_KEY")); key != "" {
		status.HasEnv = true
	}

	if key, err := LoadAPIKey(profile); err == nil && key != "" {
		status.HasKeyring = true
	}

	cfg, err := LoadConfig()
	if err == nil {
		if p, ok := cfg.Profiles[profile]; ok {
			if strings.TrimSpace(p.APIKey) != "" {
				status.HasConfig = true
			}
		}
	}

	if key, source, err := ResolveAPIKeyWithSource(profile); err == nil {
		status.Resolved = true
		status.ResolvedSource = source
		status.MaskedKey = maskKey(key)
	}

	return status
}

func maskKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	return strings.Repeat("*", len(value)-4) + value[len(value)-4:]
}
