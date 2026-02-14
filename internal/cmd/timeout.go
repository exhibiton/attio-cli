package cmd

import (
	"fmt"
	"strings"
	"time"
)

func parseTimeout(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("--timeout cannot be empty")
	}
	timeout, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid --timeout value %q: %w", raw, err)
	}
	if timeout <= 0 {
		return 0, fmt.Errorf("--timeout must be greater than 0")
	}
	return timeout, nil
}
