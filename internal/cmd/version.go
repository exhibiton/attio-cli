package cmd

import (
	"context"
	"os"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
	"github.com/failup-ventures/attio-cli/internal/ui"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

type VersionCmd struct{}

func (c *VersionCmd) Run(ctx context.Context, _ *RootFlags) error {
	payload := map[string]string{
		"version": Version,
		"commit":  Commit,
		"date":    BuildDate,
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, payload)
	}

	u := ui.FromContext(ctx)
	u.Out().Printf("version\t%s", Version)
	u.Out().Printf("commit\t%s", Commit)
	u.Out().Printf("date\t%s", BuildDate)
	return nil
}

func buildVersionString() string {
	return "attio " + Version + " (" + Commit + ", " + BuildDate + ")"
}
