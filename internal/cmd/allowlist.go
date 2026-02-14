package cmd

import (
	"fmt"
	"strings"
)

func enforceCommandAllowlist(commandPath string, allowlistRaw string) error {
	allowlist := splitCommaList(allowlistRaw)
	if len(allowlist) == 0 {
		return nil
	}

	commandPath = normalizeCommandPath(commandPath)
	if commandPath == "" {
		return nil
	}

	for _, allowed := range allowlist {
		allowed = normalizeCommandPath(allowed)
		if allowed == "" {
			continue
		}
		if commandPath == allowed || strings.HasPrefix(commandPath, allowed+" ") {
			return nil
		}
	}

	return newUsageError(fmt.Errorf("command %q is not enabled by --enable-commands=%q", commandPath, strings.Join(allowlist, ",")))
}

func normalizeCommandPath(path string) string {
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(path)))
	if len(parts) == 0 {
		return ""
	}

	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, "<") && strings.HasSuffix(part, ">") {
			continue
		}
		out = append(out, part)
	}
	return strings.Join(out, " ")
}
