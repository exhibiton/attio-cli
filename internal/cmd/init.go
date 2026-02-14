package cmd

import (
	"bufio"
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

type InitCmd struct{}

var (
	initIsTerminalFunc   = func(fd int) bool { return term.IsTerminal(fd) }
	initPromptLineFunc   = promptInitLine
	initPromptSecretFunc = promptInitSecret
	initPromptBoolFunc   = promptInitBool
)

func (c *InitCmd) Run(ctx context.Context, flags *RootFlags) error {
	if flags.NoInput {
		return newUsageError(errors.New("attio init is interactive; remove --no-input to run onboarding"))
	}
	if !initIsTerminalFunc(int(os.Stdin.Fd())) {
		return newUsageError(errors.New("attio init requires an interactive terminal"))
	}

	u := ui.FromContext(ctx)
	var promptOut io.Writer = os.Stderr
	if u != nil {
		promptOut = u.ErrWriter()
	}

	defaultProfile := config.ResolveProfile(flags.Profile)
	profile, err := initPromptLineFunc(promptOut, "Profile to configure", defaultProfile)
	if err != nil {
		return err
	}
	profile = strings.TrimSpace(profile)
	if profile == "" {
		profile = defaultProfile
	}

	apiKey, err := initPromptSecretFunc(promptOut, "Enter Attio API key")
	if err != nil {
		return err
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return newUsageError(errors.New("API key cannot be empty"))
	}

	verify, err := initPromptBoolFunc(promptOut, "Verify API key now with Attio", true)
	if err != nil {
		return err
	}
	store, err := initPromptBoolFunc(promptOut, fmt.Sprintf("Store API key in keyring for profile %q", profile), true)
	if err != nil {
		return err
	}

	baseURL := config.ResolveBaseURL(profile)
	if ok, err := maybeDryRun(ctx, "init", map[string]any{
		"profile":  profile,
		"base_url": baseURL,
		"api_key":  maskSecret(apiKey),
		"verify":   verify,
		"store":    store,
	}); ok || err != nil {
		return err
	}

	verified := false
	var self *api.Self
	if verify {
		client := api.NewClient(apiKey, baseURL)
		userAgent, timeout := getClientRuntimeOptions()
		client.SetUserAgent(userAgent)
		client.SetTimeout(timeout)

		self, err = client.GetSelf(ctx)
		if err != nil {
			return fmt.Errorf("verify API key: %w", err)
		}
		verified = true
	}

	saved := false
	if store {
		if err := config.StoreAPIKey(profile, apiKey); err != nil {
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
			"key_source":       "prompt",
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

	printInitSuccess(u, "Onboarding complete for profile %q", profile)
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
		printInitf(u, "Skipped API key verification")
	}

	if saved {
		printInitSuccess(u, "Stored API key in keyring")
	} else {
		printInitf(u, "Skipped keyring save")
	}
	printInitf(u, "Profile: %s", profile)
	printInitf(u, "Base URL: %s", baseURL)
	printInitf(u, "Next steps:")
	for _, step := range nextSteps {
		printInitf(u, "  %s", step)
	}

	return nil
}

func promptInitLine(out io.Writer, question string, defaultValue string) (string, error) {
	if defaultValue != "" {
		_, _ = fmt.Fprintf(out, "%s [%s]: ", question, defaultValue)
	} else {
		_, _ = fmt.Fprintf(out, "%s: ", question)
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue, nil
	}
	return line, nil
}

func promptInitSecret(out io.Writer, question string) (string, error) {
	_, _ = fmt.Fprintf(out, "%s: ", question)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	_, _ = fmt.Fprintln(out)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func promptInitBool(out io.Writer, question string, defaultValue bool) (bool, error) {
	label := "y/N"
	if defaultValue {
		label = "Y/n"
	}

	for {
		answer, err := promptInitLine(out, fmt.Sprintf("%s (%s)", question, label), "")
		if err != nil {
			return false, err
		}
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "":
			return defaultValue, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			_, _ = fmt.Fprintln(out, "Please answer y or n.")
		}
	}
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
