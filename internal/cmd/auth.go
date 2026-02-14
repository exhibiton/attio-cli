package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/99designs/keyring"

	"github.com/failup-ventures/attio-cli/internal/config"
	"github.com/failup-ventures/attio-cli/internal/outfmt"
	"github.com/failup-ventures/attio-cli/internal/ui"
)

type AuthCmd struct {
	Login  AuthLoginCmd  `cmd:"" help:"Store API key in keyring"`
	Logout AuthLogoutCmd `cmd:"" help:"Remove API key from keyring"`
	Status AuthStatusCmd `cmd:"" help:"Show auth status"`
}

type AuthLoginCmd struct {
	APIKey string `name:"api-key" help:"Attio API key to store. If omitted, reads from stdin when piped."`
}

func (c *AuthLoginCmd) Run(ctx context.Context, flags *RootFlags) error {
	profile := config.ResolveProfile(flags.Profile)
	apiKey := strings.TrimSpace(c.APIKey)
	if apiKey == "" {
		apiKey = readKeyFromStdin()
	}
	if apiKey == "" {
		return newUsageError(errors.New("missing API key: pass --api-key or pipe key via stdin"))
	}
	if ok, err := maybeDryRun(ctx, "auth login", map[string]any{"profile": profile, "api_key": maskSecret(apiKey)}); ok || err != nil {
		return err
	}

	if err := config.StoreAPIKey(profile, apiKey); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
			"saved":   true,
			"profile": profile,
			"source":  "keyring",
		})
	}
	u := ui.FromContext(ctx)
	u.Out().Successf("Stored API key for profile %q in keyring", profile)
	return nil
}

type AuthLogoutCmd struct{}

func (c *AuthLogoutCmd) Run(ctx context.Context, flags *RootFlags) error {
	profile := config.ResolveProfile(flags.Profile)
	if ok, err := maybeDryRun(ctx, "auth logout", map[string]any{"profile": profile}); ok || err != nil {
		return err
	}
	err := config.RemoveAPIKey(profile)
	removed := true
	if errors.Is(err, keyring.ErrKeyNotFound) {
		removed = false
		err = nil
	}
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
			"removed": removed,
			"profile": profile,
		})
	}
	u := ui.FromContext(ctx)
	if removed {
		u.Out().Successf("Removed API key for profile %q", profile)
	} else {
		u.Out().Printf("No API key stored for profile %q", profile)
	}
	return nil
}

type AuthStatusCmd struct{}

func (c *AuthStatusCmd) Run(ctx context.Context, flags *RootFlags) error {
	status := config.AuthStatus(flags.Profile)
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, status)
	}

	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "FIELD\tVALUE")
	_, _ = fmt.Fprintf(w, "profile\t%s\n", status.Profile)
	_, _ = fmt.Fprintf(w, "base_url\t%s\n", status.BaseURL)
	_, _ = fmt.Fprintf(w, "has_env\t%t\n", status.HasEnv)
	_, _ = fmt.Fprintf(w, "has_keyring\t%t\n", status.HasKeyring)
	_, _ = fmt.Fprintf(w, "has_config\t%t\n", status.HasConfig)
	_, _ = fmt.Fprintf(w, "resolved\t%t\n", status.Resolved)
	if status.Resolved {
		_, _ = fmt.Fprintf(w, "resolved_source\t%s\n", status.ResolvedSource)
		_, _ = fmt.Fprintf(w, "api_key\t%s\n", status.MaskedKey)
	}
	return nil
}

func readKeyFromStdin() string {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return ""
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return ""
	}
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
