package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/keyring"
	"golang.org/x/term"
)

const keyringServiceName = "attio-cli"

func StoreAPIKey(profile string, apiKey string) error {
	ring, err := openKeyring()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}
	return ring.Set(keyring.Item{
		Key:  keyringKey(profile),
		Data: []byte(strings.TrimSpace(apiKey)),
	})
}

func LoadAPIKey(profile string) (string, error) {
	ring, err := openKeyring()
	if err != nil {
		return "", fmt.Errorf("open keyring: %w", err)
	}
	item, err := ring.Get(keyringKey(profile))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(item.Data)), nil
}

func RemoveAPIKey(profile string) error {
	ring, err := openKeyring()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}
	return ring.Remove(keyringKey(profile))
}

func openKeyring() (keyring.Keyring, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user config dir: %w", err)
	}
	fileDir := filepath.Join(configDir, "attio-cli", "keyring")
	if err := os.MkdirAll(fileDir, 0o700); err != nil {
		return nil, fmt.Errorf("create keyring dir: %w", err)
	}

	return keyring.Open(keyring.Config{
		ServiceName: keyringServiceName,
		FileDir:     fileDir,
		FilePasswordFunc: func(_ string) (string, error) {
			return "attio-cli", nil
		},
		AllowedBackends: allowedBackends(),
	})
}

func allowedBackends() []keyring.BackendType {
	// In non-interactive/headless contexts, GUI keychain backends can block.
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return []keyring.BackendType{
			keyring.FileBackend,
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
			keyring.KWalletBackend,
			keyring.PassBackend,
		}
	}
	return []keyring.BackendType{
		keyring.KeychainBackend,
		keyring.WinCredBackend,
		keyring.SecretServiceBackend,
		keyring.KWalletBackend,
		keyring.PassBackend,
		keyring.FileBackend,
	}
}

func keyringKey(profile string) string {
	profile = ResolveProfile(profile)
	return "attio-api-key:" + profile
}
