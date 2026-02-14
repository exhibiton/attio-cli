package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type SelfCmd struct{}

func (c *SelfCmd) Run(ctx context.Context, flags *RootFlags) error {
	profile := flags.Profile
	client, err := requireClient(profile)
	if err != nil {
		return err
	}

	self, err := client.GetSelf(ctx)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, self)
	}

	w, done := tableWriter(ctx)
	defer done()

	_, _ = fmt.Fprintln(w, "FIELD\tVALUE")
	_, _ = fmt.Fprintf(w, "active\t%t\n", self.Active)
	if self.WorkspaceName != "" {
		_, _ = fmt.Fprintf(w, "workspace_name\t%s\n", self.WorkspaceName)
	}
	if self.WorkspaceSlug != "" {
		_, _ = fmt.Fprintf(w, "workspace_slug\t%s\n", self.WorkspaceSlug)
	}
	if self.WorkspaceID != "" {
		_, _ = fmt.Fprintf(w, "workspace_id\t%s\n", self.WorkspaceID)
	}
	if self.Scope != "" {
		_, _ = fmt.Fprintf(w, "scope\t%s\n", self.Scope)
	}
	if self.ClientID != "" {
		_, _ = fmt.Fprintf(w, "client_id\t%s\n", self.ClientID)
	}
	if self.IAT != nil {
		_, _ = fmt.Fprintf(w, "iat\t%d\n", *self.IAT)
	}
	if self.Exp != nil {
		_, _ = fmt.Fprintf(w, "exp\t%d\n", *self.Exp)
	}

	return nil
}
