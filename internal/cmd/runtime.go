package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
	"github.com/failup-ventures/attio-cli/internal/ui"
)

type RuntimeOptions struct {
	DryRun    bool
	FailEmpty bool
	IDOnly    bool
}

type runtimeOptionsKey struct{}

func withRuntimeOptions(ctx context.Context, opts RuntimeOptions) context.Context {
	return context.WithValue(ctx, runtimeOptionsKey{}, opts)
}

func runtimeOptionsFromContext(ctx context.Context) RuntimeOptions {
	v := ctx.Value(runtimeOptionsKey{})
	if v == nil {
		return RuntimeOptions{}
	}
	opts, _ := v.(RuntimeOptions)
	return opts
}

func isDryRun(ctx context.Context) bool {
	return runtimeOptionsFromContext(ctx).DryRun
}

func isFailEmpty(ctx context.Context) bool {
	return runtimeOptionsFromContext(ctx).FailEmpty
}

func isIDOnly(ctx context.Context) bool {
	return runtimeOptionsFromContext(ctx).IDOnly
}

func maybeDryRun(ctx context.Context, action string, payload any) (bool, error) {
	if !isDryRun(ctx) {
		return false, nil
	}

	if outfmt.IsJSON(ctx) {
		return true, outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
			"dry_run": true,
			"action":  action,
			"data":    payload,
		})
	}

	u := ui.FromContext(ctx)
	if u != nil {
		u.Out().Printf("[dry-run] %s", action)
		if payload != nil {
			u.Out().Printf("[dry-run] payload: %s", anyString(payload))
		}
	} else {
		_, _ = os.Stdout.WriteString("[dry-run] " + action + "\n")
		if payload != nil {
			_, _ = os.Stdout.WriteString("[dry-run] payload: " + anyString(payload) + "\n")
		}
	}
	return true, nil
}

func maybeFailEmpty(ctx context.Context, count int) error {
	if count > 0 || !isFailEmpty(ctx) {
		return nil
	}
	return &ExitError{Code: ExitCodeNoResult, Err: errors.New("no results")}
}

func maybePreviewResults(ctx context.Context, count int) error {
	if err := maybeFailEmpty(ctx, count); err != nil {
		if outfmt.IsJSON(ctx) {
			return err
		}
		u := ui.FromContext(ctx)
		if u != nil {
			u.Err().Error("No results")
		} else {
			_, _ = fmt.Fprintln(os.Stderr, "No results")
		}
		return err
	}
	return nil
}
