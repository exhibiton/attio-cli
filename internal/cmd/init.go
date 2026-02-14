package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/failup-ventures/attio-cli/internal/api"
	"github.com/failup-ventures/attio-cli/internal/config"
	"github.com/failup-ventures/attio-cli/internal/outfmt"
	"github.com/failup-ventures/attio-cli/internal/ui"
)

type InitCmd struct {
	APIKey     string `name:"api-key" help:"Attio API key. If omitted, uses stdin, ATTIO_API_KEY, or interactive prompt."`
	NoStore    bool   `name:"no-store" help:"Verify but do not store API key in keyring"`
	SkipVerify bool   `name:"skip-verify" help:"Skip API key verification against /v2/self"`
}

func (c *InitCmd) Run(ctx context.Context, flags *RootFlags) error {
	profile := config.ResolveProfile(flags.Profile)
	key, source, err := resolveInitAPIKey(c.APIKey, flags.NoInput, ui.FromContext(ctx))
	if err != nil {
		return err
	}
	if key == "" {
		return newUsageError(errors.New("missing API key: pass --api-key, set ATTIO_API_KEY, or pipe key via stdin"))
	}

	baseURL := config.ResolveBaseURL(profile)
	if ok, err := maybeDryRun(ctx, "init", map[string]any{
		"profile":     profile,
		"base_url":    baseURL,
		"key_source":  source,
		"api_key":     maskSecret(key),
		"skip_verify": c.SkipVerify,
		"no_store":    c.NoStore,
	}); ok || err != nil {
		return err
	}

	verified := !c.SkipVerify
	var self *api.Self
	if verified {
		client := api.NewClient(key, baseURL)
		userAgent, timeout := getClientRuntimeOptions()
		client.SetUserAgent(userAgent)
		client.SetTimeout(timeout)

		self, err = client.GetSelf(ctx)
		if err != nil {
			return fmt.Errorf("verify API key: %w", err)
		}
	}

	saved := false
	if !c.NoStore {
		if err := config.StoreAPIKey(profile, key); err != nil {
			return err
		}
		saved = true
	}

	nextSteps := []string{
		fmt.Sprintf("attio --profile %s auth status --json", profile),
		fmt.Sprintf("attio --profile %s self", profile),
		fmt.Sprintf("attio --profile %s objects list", profile),
	}

	if outfmt.IsJSON(ctx) {
		payload := map[string]any{
			"initialized":      true,
			"profile":          profile,
			"base_url":         baseURL,
			"key_source":       source,
			"saved_to_keyring": saved,
			"verified":         verified,
			"next_steps":       nextSteps,
		}
		if self != nil {
			payload["workspace_name"] = self.WorkspaceName
			payload["workspace_slug"] = self.WorkspaceSlug
			payload["workspace_id"] = self.WorkspaceID
		}
		return outfmt.WriteJSON(ctx, os.Stdout, payload)
	}

	u := ui.FromContext(ctx)
	if verified {
		workspaceLabel := strings.TrimSpace(self.WorkspaceName)
		if workspaceLabel == "" {
			workspaceLabel = strings.TrimSpace(self.WorkspaceSlug)
		}
		if workspaceLabel != "" {
			printInitSuccess(u, "Verified API key for workspace %q", workspaceLabel)
		} else {
			printInitSuccess(u, "Verified API key successfully")
		}
	} else {
		printInitf(u, "Skipped API key verification (--skip-verify)")
	}

	if saved {
		printInitSuccess(u, "Stored API key for profile %q in keyring", profile)
	} else {
		printInitf(u, "Skipped keyring save (--no-store)")
	}
	printInitf(u, "Profile: %s", profile)
	printInitf(u, "Base URL: %s", baseURL)
	printInitf(u, "Key source: %s", source)
	printInitf(u, "Next steps:")
	for _, step := range nextSteps {
		printInitf(u, "  %s", step)
	}

	return nil
}

func resolveInitAPIKey(explicit string, noInput bool, u *ui.UI) (string, string, error) {
	if key := strings.TrimSpace(explicit); key != "" {
		return key, "flag", nil
	}
	if key := readKeyFromStdin(); key != "" {
		return key, "stdin", nil
	}
	if key := strings.TrimSpace(os.Getenv("ATTIO_API_KEY")); key != "" {
		return key, "env", nil
	}
	if noInput {
		return "", "", nil
	}

	key, err := readKeyFromPrompt(u)
	if err != nil {
		return "", "", err
	}
	if key != "" {
		return key, "prompt", nil
	}
	return "", "", nil
}

func readKeyFromPrompt(u *ui.UI) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", nil
	}

	var out io.Writer = os.Stderr
	if u != nil {
		out = u.ErrWriter()
	}

	_, _ = fmt.Fprint(out, "Enter Attio API key: ")
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	_, _ = fmt.Fprintln(out)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func printInitSuccess(u *ui.UI, format string, args ...any) {
	if u != nil {
		u.Out().Successf(format, args...)
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}

func printInitf(u *ui.UI, format string, args ...any) {
	if u != nil {
		u.Out().Printf(format, args...)
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}
